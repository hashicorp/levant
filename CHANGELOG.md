## 0.0.3 (Unreleased)

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
