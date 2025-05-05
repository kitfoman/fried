# gpuFryer

A simple daemonset that runs in a Kubernetes cluster and executes DCGM diagnostics on demand for GPU burn-in tests before putting GPUs into service. The server currently supports two endpoints:
- A scheduling endpoint to initiate a diagnostic on a node targeting specific GPUs at a given diagnostic level, which returns an ID for the routine that starts the diagnostics. Returns an error if its unable to schedule the job at all.

- A status endpoint that returns the current status and/or result of the diagnostic
