# Contributing to Levant

**First** of all welcome and thank you for even considering to contribute to the Levant project. If you're unsure or afraid of anything, just ask or submit the issue or pull request anyways. You won't be yelled at for giving your best effort.

If you wish to work on Levant itself or any of its built-in components, you will first need [Go](https://golang.org/) installed on your machine (version 1.9+ is required) and ensure your [GOPATH](https://golang.org/doc/code.html#GOPATH) is correctly configured.

## Issues

Remember to craft your issues in such a way as to help others who might be facing similar challenges. Give your issues meaningful titles, that offer context. Please try to use complete sentences in your issue. Everything is okay to submit as an issue, even questions or ideas.

Not every contribution requires an issue, but all bugs and significant changes to core functionality do. A pull request to fix a bug or implement a core change will not be accepted without a corresponding bug report.

## What Good Issues Look Like

Levant includes a default issue template, please endeavor to provide as much information as possible.

1. **Avoid raising duplicate issues.** Please use the GitHub issue search feature to check whether your bug report or feature request has been mentioned in the past.

1. **Provide Bug Details.** When filing a bug report, include debug logs, version details and stack traces. Your issue should provide:
    1. Guidance on **how to reproduce the issue.**
    1. Tell us **what you expected to happen.**
    1. Tell us **what actually happened.**
    1. Tell us **what version of Levant and Nomad you're using.**

1. **Provide Feature Details.** When filing a feature request, include background on why you're requesting the feature and how it will be useful to others. If you have a design proposal to implement the feature, please include these details so the maintainers can provide feedback on the proposed approach.    

## Pull Requests

**All pull requests must include a description.** The description should at a minimum, provide background on the purpose of the pull request. Consider providing an overview of why the work is taking place; don’t assume familiarity with the history. If the pull request is related to an issue, make sure to mention the issue number(s).

Try to keep pull requests tidy, and be prepared for feedback. Everyone is welcome to contribute to Levant but we do try to keep a high quality of code standard. Be ready to face this. Feel free to open a pull request for anything, about anything. **Everyone** is welcome.

## Get Early Feedback

If you are contributing, do not feel the need to sit on your contribution until it is perfectly polished and complete. It helps everyone involved for you to seek feedback as early as you possibly can. Submitting an early, unfinished version of your contribution for feedback in no way prejudices your chances of getting that contribution accepted, and can save you from putting a lot of work into a contribution that is not suitable for the project.

## Code Review

Pull requests will not be merged until they've been code reviewed by at least one maintainer. You should implement any code review feedback unless you strongly object to it. In the event that you object to the code review feedback, you should make your case clearly and calmly. If, after doing so, the feedback is judged to still apply, you must either apply the feedback or withdraw your contribution.

# Tests and Checks

As Levant continues to mature, additional test harnesses will be implemented. Once these harnesses are in place, tests will be required for all bugs fixes and features. No exception.

## Linting

All Go code in your pull request must pass `lint` checks. You can run lint on all Golang files using the `make lint` target. All lint checks are automatically enforced by CI tests.

## Formatting

**Do your best to follow existing conventions you see in the codebase**, and ensure your code is formatted with `go fmt`. You can run `fmt` on all Golang files using the `make fmt` target. All format checks are automatically enforced by CI tests.

## Testing

Tests are required for all new functionality where practical; certain portions of Levant have no tests but this should be the exception and not the rule.

# Building

Levant is linted, tested and built using make:

```
make
```

The resulting binary file will be stored in the project root directory and is named `Levant-local` which can be invoked as required. The binary is built by default for the host system only. You can cross-compile and build binaries for a number of different systems and architectures by invoking the build script:

```
./scripts/build.sh
```

The build script outputs the binary files to `/pkg`:

```
darwin-386-levant
darwin-amd64-levant
freebsd-386-levant
freebsd-amd64-levant
freebsd-arm-levant
linux-386-levant
linux-amd64-levant
linux-arm-levant
```

See [docs](https://golang.org/doc/install/source) for the whole list of available `GOOS` and `GOARCH`
values.
