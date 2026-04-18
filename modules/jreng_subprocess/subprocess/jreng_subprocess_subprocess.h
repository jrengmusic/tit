#pragma once

#include <functional>
#include <memory>

namespace jreng
{ /*____________________________________________________________________________*/

/*____________________________________________________________________________*/
/** Launches a child process, streams output chunks to the caller, and delivers
    a single completion callback with captured stdout + stderr when the process
    exits or is killed.

    Threading contract:
    - `launch()` and `kill()` are called from the message thread.
    - `Handler::Chunk` callbacks fire on the Worker thread.
      Callers that need message-thread delivery must marshal via
      `juce::MessageManager::callAsync` at the consumer layer.
    - `Handler::Completion` fires on the Worker thread immediately after the
      process exits. Same marshaling responsibility applies.

    Byte-cap contract:
    - stdout and stderr are each capped at `BYTE_CAP` bytes.
    - When the cap is reached, `TRUNCATION_NOTICE` is appended once and reading
      for that stream stops. The process continues running.

    Ownership:
    - `Subprocess` owns the `Worker` exclusively via `std::unique_ptr`.
    - Destroying `Subprocess` while a process is running calls `kill()` first.

    Usage:
    @code
    jreng::Subprocess sub;
    sub.launch ({"git", "fetch", "--progress"},
                juce::File ("/path/to/repo"),
                [](int code, const juce::String& out, const juce::String& err)
                {
                    // completion — marshal to message thread if needed
                },
                [](juce::StringRef chunk, bool isReplace)
                {
                    // streaming chunk — marshal to message thread if needed
                });
    @endcode

    @see Handler
*/
class Subprocess
{
public:
    //==========================================================================
    /** Callback type grouping for Subprocess.

        Mirrors the `jreng::File::Watcher` nested-type precedent.
        Parameter names using `on*` prefix at call sites are NAMES-compliant
        for callback-invocation parameters.
    */
    struct Handler
    {
        /** Called once when the process exits or `kill()` is called.

            @param exitCode       Process exit code (platform-defined; 0 = success).
            @param stdoutCapture  Accumulated stdout, truncated at `BYTE_CAP` bytes.
            @param stderrCapture  Accumulated stderr, truncated at `BYTE_CAP` bytes.
        */
        using Completion = std::function<void (int exitCode,
                                               const juce::String& stdoutCapture,
                                               const juce::String& stderrCapture)>;

        /** Called for each output chunk delivered from the process.

            @param chunk      Raw bytes read from stdout or stderr.
            @param isReplace  True when the chunk ends with `\r` (in-place progress
                              update, LogReplace-equivalent). False when `\n`-terminated
                              (append, Log-equivalent).
        */
        using Chunk = std::function<void (juce::StringRef chunk, bool isReplace)>;
    };

    //==========================================================================
    Subprocess();
    ~Subprocess();

    //==========================================================================
    /** Launches a subprocess.

        Noop if a process is already running — `kill()` first if replacement is needed.

        @param command     Pre-tokenized argv. Index 0 is the executable name.
                           Matches Go TIT's `exec.Command(name, args...)`.
        @param workingDir  Working directory for the child process.
        @param onComplete  Called once when the process exits. May be nullptr.
        @param onChunk     Called for each output chunk. nullptr = capture-only mode.
    */
    void launch (const juce::StringArray& command,
                 const juce::File& workingDir,
                 Handler::Completion onComplete,
                 Handler::Chunk onChunk = nullptr);

    /** Terminates the running process and waits for the Worker thread to exit.

        Safe to call when no process is running.
    */
    void kill();

    //==========================================================================
    /** Maximum bytes accumulated per stream before truncation. */
    static constexpr int BYTE_CAP { 100'000 };

    /** Appended to a stream's accumulation buffer when `BYTE_CAP` is reached. */
    static constexpr const char* TRUNCATION_NOTICE { "\n[output truncated — byte cap reached]" };

    /** Environment variable suppressing interactive git terminal prompts. */
    static constexpr const char* ENV_TERMINAL_PROMPT { "GIT_TERMINAL_PROMPT=0" };

    /** Environment variable suppressing git progress delay. */
    static constexpr const char* ENV_PROGRESS_DELAY { "GIT_PROGRESS_DELAY=0" };

private:
    class Worker;

    std::unique_ptr<Worker> worker;

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (Subprocess)
};

/**_____________________________END OF NAMESPACE______________________________*/
}// namespace jreng
