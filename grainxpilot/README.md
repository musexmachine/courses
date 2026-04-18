# grainxpilot Go bundle

Standalone Go bundle for the Grain -> normalize -> batch folder -> browser -> X-Pilot pipeline.

Included:
- typed config and run/item state enums
- Grain adapter interface
- normalizer and browser worker interfaces
- manifest validation
- run-layout helpers
- stub worker implementations returning `ErrNotImplemented`

This bundle is repo-agnostic because the target repository was not reachable from this runtime.
Adjust the module path and folder placement when landing it in the real repo.
