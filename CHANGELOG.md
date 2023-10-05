## UNRELEASED

## 0.3.3 (October 5, 2023)

IMPROVEMENTS:
* build: Now builds with Go v1.21.1 [[GH-507](https://github.com/hashicorp/levant/pull/507)]
* deps: Updated Nomad dependency to v1.6.2 [[GH-507](https://github.com/hashicorp/levant/pull/507)]

## 0.3.2 (October 20, 2022)

IMPROVEMENTS:
 * build: Now builds with go v1.19.1 [[GH-464](https://github.com/hashicorp/levant/pull/464)]
 * deps: Updated Nomad dependency to v1.4.1. [[GH-464](https://github.com/hashicorp/levant/pull/464)]
 * deps: Updated golang.org/x/text to v0.4.0 [[GH-465](https://github.com/hashicorp/levant/pull/465)]

## 0.3.1 (February 14, 2022)

IMPROVEMENTS:
* build: Updated Nomad dependency to 1.2.4. [[GH-438](https://github.com/hashicorp/levant/pull/438)]

## 0.3.0 (March 09, 2021)

__BACKWARDS INCOMPATIBILITIES:__
 * template: existing Levant functions that share a name with [sprig](https://github.com/Masterminds/sprig) functions have been renamed to include the prefix `levant` such as `levantEnv`.

BUG FIXES:
 * cli: Fixed panic when dispatching a job. [[GH-348](https://github.com/hashicorp/levant/pull/348)]
 * status-checker: Pass the namespace to the query options when calling the Nomad API [[GH-356](https://github.com/hashicorp/levant/pull/356)]
 * template: Fixed issue with default variables file not being used. [[GH-353](https://github.com/hashicorp/levant/pull/353)]

IMPROVEMENTS:
 * build: Updated Nomad dependency to 1.0.4. [[GH-399](https://github.com/hashicorp/levant/pull/399)]
 * cli: Added `log-level` and `log-format` flags to render command. [[GH-346](https://github.com/hashicorp/levant/pull/346)]
 * render: when rendering, send logging to stderr if stdout is not a terminal [[GH-386](https://github.com/hashicorp/levant/pull/386)]
 * template: Added [sprig](https://github.com/Masterminds/sprig) template functions. [[GH-347](https://github.com/hashicorp/levant/pull/347)]
 * template: Added `spewDump` and `spewPrintf` functions for easier debugging. [[GH-344](https://github.com/hashicorp/levant/pull/344)]

## 0.2.9 (27 December 2019)

IMPROVEMENTS:
 * Update vendoered version of Nomad to 0.9.6 [GH-313](https://github.com/jrasell/levant/pull/313)
 * Update to go 1.13 and use modules rather than dep [GH-319](https://github.com/jrasell/levant/pull/319)
 * Remove use of vendor nomad/structs import to allow easier vendor [GH-320](https://github.com/jrasell/levant/pull/320)
 * Add template replace function [GH-291](https://github.com/jrasell/levant/pull/291)

BUG FIXES:
 * Use info level logs when no changes are detected [GH-303](https://github.com/jrasell/levant/pull/303)

## 0.2.8 (14 September 2019)

IMPROVEMENTS:
 * Add `-force` flag to deploy CLI command which allows for forcing a deployment even if Levant detects 0 changes on plan [GH-296](https://github.com/jrasell/levant/pull/296)

BUG FIXES:
 * Fix segfault when logging deployID details [GH-286](https://github.com/jrasell/levant/pull/286)
 * Fix error message within scale-in which incorrectly referenced scale-out [GH-285](https://github.com/jrasell/levant/pull/285/files)

## 0.2.7 (19 March 2019)

IMPROVEMENTS:
 * Use `missingkey=zero` rather than error which allows better use of standard go templating, particulary conditionals [GH-275](https://github.com/jrasell/levant/pull/275)
 * Added maths functions add, subtract, multiply, divide and modulo to the template rendering process [GH-277](https://github.com/jrasell/levant/pull/277)

## 0.2.6 (25 February 2019)

IMPROVEMENTS:
 * Add the ability to supply a Vault token to a job during deployment via either a `vault` or `vault-token` flag [GH-258](https://github.com/jrasell/levant/pull/258)
 * New `fileContents` template function which allows the entire contents of a file to be read into the template [GH-261](https://github.com/jrasell/levant/pull/261)

BUG FIXES:
 * Fix a panic when running scale* deployment watcher due to incorrectly initialized client config [GH-253](https://github.com/jrasell/levant/pull/253)
 * Fix incorrect behavior when flag `ignore-no-changes` was set [GH-264](https://github.com/jrasell/levant/pull/264)
 * Fix endless deployment loop when Nomad doesn't return a deployment ID [GH-268](https://github.com/jrasell/levant/pull/268)

## 0.2.5 (25 October 2018)

BUG FIXES:
 * Fix panic in deployment where count is not specified due to unsafe count checking on task groups [GH-249](https://github.com/jrasell/levant/pull/249)

## 0.2.4 (24 October 2018)

BUG FIXES:
 * Fix panic in scale commands due to an incorrectly initialized configuration struct [GH-244](https://github.com/jrasell/levant/pull/244)
 * Fix bug where job deploys with taskgroup counts of 0 would hang for 1 hour [GH-246](https://github.com/jrasell/levant/pull/246)

## 0.2.3 (2 October 2018)

IMPROVEMENTS:
 * New `env` template function allows the lookup and substitution of variables by environemnt variables [GH-225](https://github.com/jrasell/levant/pull/225)
 * Add plan command to allow running a plan whilst using templating [GH-234](https://github.com/jrasell/levant/pull/234)
 * Add `toUpper` and `toLower` template funcs [GH-237](https://github.com/jrasell/levant/pull/237)

## 0.2.2 (6 August 2018)

BUG FIXES:
 * Fix an issue where if an evaluation had filtered nodes Levant would exit immediately rather than tracking the deployment which could still succeed [GH-221](https://github.com/jrasell/levant/pull/221)
 * Fixed failure inspector to report on tasks that are restarting [GH-82](https://github.com/jrasell/levant/pull/82)

## 0.2.1 (20 July 2018)

IMPROVEMENTS:
 * JSON can now be used as a variable file format [GH-210](https://github.com/jrasell/levant/pull/210)
 * The template funcs now include numerous parse functions to provide greater flexibility [GH-212](https://github.com/jrasell/levant/pull/212)
 * Ability to configure allow-stale Nomad setting when performing calls to help in environments with high network latency [GH-185](https://github.com/jrasell/levant/pull/185)
 
BUG FIXES:
 * Update vendored package of Nomad to fix failures when interacting with jobs configured with update progress_deadline params [GH-216](https://github.com/jrasell/levant/pull/216)

## 0.2.0 (4 July 2018)

IMPROVEMENTS:
 * New `scale-in` and `scale-out` commands  allow an operator to manually scale jobs and task groups based on counts or percentages [GH-172](https://github.com/jrasell/levant/pull/172)
 * New template functions allowing the lookup of variables from Consul KVs, ISO-8601 timestamp generation and loops [GH-175](https://github.com/jrasell/levant/pull/175), [GH-202](https://github.com/jrasell/levant/pull/202)
 * Multiple variable files can be passed on each run, allowing for common configuration to be shared across jobs [GH-180](https://github.com/jrasell/levant/pull/180)
 * Provide better command help for deploy and render commands [GH-183](https://github.com/jrasell/levant/pull/184)
 * Add `-ignore-no-changes` flag to deploy CLI command which allows the changing on behaviour to exit 0 even if Levant detects 0 changes on plan [GH-196](https://github.com/jrasell/levant/pull/196)

BUG FIXES:
 * Fix formatting with version summary output which had erronous quote [GH-170](https://github.com/jrasell/levant/pull/170)

## 0.1.1 (13 May 2018)

IMPROVEMENTS:
 * Use govvv for builds and to supply additional version information in the version command output [GH-151](https://github.com/jrasell/levant/pull/151)
 * Levant will now run Nomad plan before deployments to log the plan diff [GH-153](https://github.com/jrasell/levant/pull/153)
 * Logging can now be output in JSON format and uses contextual data for better processing ability [GH-157](https://github.com/jrasell/levant/pull/157)
 
BUG FIXES:
 * Fix occasional panic when performing deployment check of a batch job deployment [GH-150](https://github.com/jrasell/levant/pull/150)

## 0.1.0 (18 April 2018)

IMPROVEMENTS:
 * New 'dispatch' command which allows Levant to dispatch Nomad jobs which will go through Levants additional job checking [GH-128](https://github.com/jrasell/levant/pull/128)
 * New 'force-batch' deploy flag which allows users to trigger a periodic run on deployment independent of the schedule [GH-110](https://github.com/jrasell/levant/pull/110) 
 * Enhanced job status checking for non-service type jobs [GH-96](https://github.com/jrasell/levant/pull/96), [GH-109](https://github.com/jrasell/levant/pull/109)
 * Implement config struct for Levant to track config during run [GH-102](https://github.com/jrasell/levant/pull/102)
 * Test and build Levant with Go version 1.10 [GH-119](https://github.com/jrasell/levant/pull/119), [GH-116](https://github.com/jrasell/levant/pull/116)
 * Add a catchall for unhandled failure cases to log more useful information for the operator [GH-138](https://github.com/jrasell/levant/pull/138)
 * Updated vendored dependancy of Nomad to 0.8.0 [GH-137](https://github.com/jrasell/levant/pull/137)
 
BUG FIXES:
 * Service jobs that don't have an update stanza do not produce deployments and should skip the deployment watcher [GH-99](https://github.com/jrasell/levant/pull/99)
 * Ensure the count updater ignores jobs that are in stopped state [GH-106](https://github.com/jrasell/levant/pull/106)
 * Fix a small formatting issue with the deploy command arg help [GH-111](https://github.com/jrasell/levant/pull/111)
 * Do not run the auto-revert inspector if auto-promote fails [GH-122](https://github.com/jrasell/levant/pull/122)
 * Fix issue where allocationStatusChecker logged incorrectly [GH-131](https://github.com/jrasell/levant/pull/131)
 * Add retry to auto-revert checker to ensure the correct deployment is monitored, and not the original [GH-134](https://github.com/jrasell/levant/pull/134)

## 0.0.4 (25 January 2018)

IMPROVEMENTS:
 * Job types of `batch` now undergo checking to confirm the job reaches status of `running` [GH-73](https://github.com/jrasell/levant/pull/73)
 * Vendored Nomad version has been increased to 0.7.1 allowing use of Nomad ACL tokens [GH-76](https://github.com/jrasell/levant/pull/76)
 * Log messages now includes the date, time and timezone [GH-80](https://github.com/jrasell/levant/pull/80)

BUG FIXES:
 * Skip health checks for task groups without canaries when performing canary auto-promote health checking [GH-83](https://github.com/jrasell/levant/pull/83)
 * Fix issue where jobs without specified count caused panic [GH-89](https://github.com/jrasell/levant/pull/89)

## 0.0.3 (23 December 2017)

IMPROVEMENTS:
 * Levant can now track Nomad auto-revert of a failed deployment [GH-55](https://github.com/jrasell/levant/pull/55)
 * Provide greater feedback around variables file passed, CLI variables passed and which variables are being used by Levant.[GH-62](https://github.com/jrasell/levant/pull/62)
 * Levant supports autoloading of default files when running `levant deploy` [GH-37](https://github.com/jrasell/levant/pull/37)

BUG FIXES:
 * Fix issue where Levant did not correctly handle deploying jobs of type `batch` [GH-52](https://github.com/jrasell/levant/pull/52)
 * Fix issue where evaluations errors were not being fully checked [GH-66](https://github.com/jrasell/levant/pull/66)
 * Fix issue in failure_inspector incorrectly handling multi-groups [GH-69](https://github.com/jrasell/levant/pull/69)

## 0.0.2 (29 November 2017)

IMPROVEMENTS:
 * Introduce `-force-count` flag into deploy command which disables dynamic count updating; meaning Levant will explicity use counts defined in the job specification template [GH-33](https://github.com/jrasell/levant/pull/33)
 * Levant deployments now inspect the evaluation results and log any error messages [GH-40](https://github.com/jrasell/levant/pull/40)

BUG FIXES:
 * Fix formatting issue in render command help [GH-28](https://github.com/jrasell/levant/pull/28)
 * Update failure_inspector to cover more failure use cases [GH-27](https://github.com/jrasell/levant/pull/27)
 * Fix a bug in handling Nomad job types incorrectly [GH-32](https://github.com/jrasell/levant/pull/32)
 * Fix issue where jobs deployed with all task group counts at 0 would cause a failure as no deployment ID is returned [GH-36](https://github.com/jrasell/levant/pull/36)

## 0.0.1 (30 October 2017)

- Initial release.
