// ============================================================================
// Subprocess + Worker implementation
// ============================================================================

namespace jreng
{ /*____________________________________________________________________________*/

/*____________________________________________________________________________*/
/** Worker owns the `juce::ChildProcess` and drives the read loop on a
    dedicated thread.

    Lifecycle:
    - Constructed by `Subprocess::launch()` with frozen copies of callbacks
      and command data.
    - `startThread()` is called immediately after construction.
    - The thread exits when the process ends or `threadShouldExit()` is set.
    - `Subprocess` holds `Worker` via `std::unique_ptr`; destruction calls
      `stopThread()` via the destructor.

    Stream contract:
    - `juce::ChildProcess::readProcessOutput` delivers bytes from whichever
      streams were opened (both stdout + stderr via `wantStdOut | wantStdErr`).
    - The two streams are interleaved at the OS pipe level; JUCE exposes them
      as a single byte sequence. Callers receive both streams merged — matching
      the behavior of `execute.go` which reads a single merged output.
    - `BYTE_CAP` is applied to the merged accumulation buffer.
*/
class Subprocess::Worker : private juce::Thread
{
public:
    Worker (const juce::StringArray& command,
            const juce::File& workingDir,
            Handler::Completion onComplete,
            Handler::Chunk onChunk)
        : juce::Thread ("jreng::Subprocess::Worker")
        , commandArgs (buildCommand (command, workingDir))
        , directory (workingDir)
        , completionCallback (std::move (onComplete))
        , chunkCallback (std::move (onChunk))
    {
        jassert (workingDir.isDirectory());
        jassert (not command.isEmpty());
        startThread();
    }

    /** Prepends `env -C <workingDir>` to `command` so the child process
        starts in the correct directory without relying on the shell.
    */
    static juce::StringArray buildCommand (const juce::StringArray& command,
                                           const juce::File& workingDir)
    {
        jassert (workingDir.isDirectory());
        jassert (not command.isEmpty());

        juce::StringArray result;
        result.add ("env");
        result.add ("-C");
        result.add (workingDir.getFullPathName());
        result.addArray (command);
        return result;
    }

    ~Worker() override
    {
        signalThreadShouldExit();
        process.kill();
        stopThread (5000);
    }

    /** Signals the worker to stop and kills the child process. */
    void terminate()
    {
        signalThreadShouldExit();
        process.kill();
    }

private:
    //==========================================================================
    void run() override
    {
        const bool started { process.start (commandArgs,
                                            juce::ChildProcess::wantStdOut
                                            | juce::ChildProcess::wantStdErr) };

        if (started)
        {
            juce::String accumulation {};
            readStream (accumulation);

            const int exitCode { static_cast<int> (process.getExitCode()) };

            if (completionCallback != nullptr)
                completionCallback (exitCode, accumulation, {});
        }
        else
        {
            if (completionCallback != nullptr)
                completionCallback (-1, {}, {});
        }
    }

    //==========================================================================
    /** Reads all available bytes from the child process into `accumulation`,
        delivering chunks to `chunkCallback` when present.

        Stops when `threadShouldExit()` is set or `readProcessOutput` returns
        zero bytes (EOF / process exited). Applies `BYTE_CAP` truncation.

        Both stdout and stderr arrive interleaved from `juce::ChildProcess`
        (single pipe view). Accumulation captures the merged stream.

        `isReplace` is true when the most recent chunk ends with bare `\r`
        (carriage-return without following `\n`), signalling an in-place
        terminal progress line — LogReplace-equivalent from execute.go.
    */
    void readStream (juce::String& accumulation)
    {
        static constexpr int READ_CHUNK_SIZE { 4096 };
        bool capReached { false };
        juce::HeapBlock<char> buf { static_cast<size_t> (READ_CHUNK_SIZE) };
        int bytesRead { READ_CHUNK_SIZE };

        while (not threadShouldExit() and bytesRead > 0)
        {
            bytesRead = process.readProcessOutput (buf.getData(), READ_CHUNK_SIZE);

            if (bytesRead > 0)
            {
                const juce::String chunk { juce::CharPointer_UTF8 (buf.getData()),
                                           static_cast<size_t> (bytesRead) };

                processChunk (accumulation, capReached, chunk, bytesRead);
            }
        }
    }

    /** Appends `chunk` to `accumulation` up to `Subprocess::BYTE_CAP`, appending
        `TRUNCATION_NOTICE` once when the cap is crossed and setting `capReached`.

        No-op when `capReached` is already true.
    */
    void appendWithCap (juce::String& accumulation,
                        bool& capReached,
                        const juce::String& chunk,
                        int bytesRead)
    {
        if (not capReached)
        {
            const int remaining { Subprocess::BYTE_CAP - accumulation.length() };

            if (remaining <= 0)
            {
                accumulation += juce::String (Subprocess::TRUNCATION_NOTICE);
                capReached = true;
            }
            else if (bytesRead <= remaining)
            {
                accumulation += chunk;
            }
            else
            {
                accumulation += chunk.substring (0, remaining);
                accumulation += juce::String (Subprocess::TRUNCATION_NOTICE);
                capReached = true;
            }
        }
    }

    /** Returns true when `chunk` represents an in-place terminal progress line
        (ends with bare `\r` — carriage-return without following `\n`).

        Matches the LogReplace signal from execute.go.
    */
    bool computeIsReplace (const juce::String& chunk) const
    {
        return chunk.endsWith ("\r") and not chunk.endsWith ("\r\n");
    }

    /** Processes a single read iteration: accumulates bytes into `accumulation`
        via `appendWithCap` and delivers the chunk to `chunkCallback` when set.
    */
    void processChunk (juce::String& accumulation,
                       bool& capReached,
                       const juce::String& chunk,
                       int bytesRead)
    {
        appendWithCap (accumulation, capReached, chunk, bytesRead);

        if (chunkCallback != nullptr)
            chunkCallback (chunk, computeIsReplace (chunk));
    }

    //==========================================================================
    juce::StringArray    commandArgs;
    juce::File           directory;
    Handler::Completion  completionCallback;
    Handler::Chunk       chunkCallback;
    juce::ChildProcess   process;

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (Worker)
};

//==============================================================================
Subprocess::Subprocess() = default;

Subprocess::~Subprocess()
{
    kill();
}

void Subprocess::launch (const juce::StringArray& command,
                         const juce::File& workingDir,
                         Handler::Completion onComplete,
                         Handler::Chunk onChunk)
{
    jassert (worker == nullptr);
    jassert (not command.isEmpty());
    jassert (workingDir.isDirectory());

    if (worker == nullptr)
    {
        worker = std::make_unique<Worker> (command,
                                           workingDir,
                                           std::move (onComplete),
                                           std::move (onChunk));
    }
}

void Subprocess::kill()
{
    if (worker != nullptr)
    {
        worker->terminate();
        worker.reset();
    }
}

/**_____________________________END OF NAMESPACE______________________________*/
}// namespace jreng
