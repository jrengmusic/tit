#include <juce_core/juce_core.h>

int main (int, char**)
{
    juce::UnitTestRunner runner {};
    runner.runAllTests ();

    int numFailed {0};

    for (int i {0}; i < runner.getNumResults (); ++i)
    {
        const juce::UnitTestRunner::TestResult* result {runner.getResult (i)};

        if (result not_eq nullptr)
        {
            juce::String status {result->failures > 0 ? "[FAIL]" : "[PASS]"};
            juce::String line {status + " " + result->unitTestName + " / " + result->subcategoryName
                               + " — failures: " + juce::String (result->failures)
                               + ", passes: " + juce::String (result->passes)};

            std::cout << line.toStdString () << "\n";

            if (result->failures > 0)
            {
                ++numFailed;
            }
        }
    }

    if (runner.getNumResults () == 0)
    {
        std::cout << "No tests were registered.\n";
    }

    return numFailed > 0 ? 1 : 0;
}
