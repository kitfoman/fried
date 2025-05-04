# fried

Fried is a set of workflows designed to burn in GPUs before putting them into the service. It runs a set of tests/benchmarks to test
the GPUs, the interconnect, PCIE tests and cpu tests. 

Burn in tests:
 - Linpack via nvcr.io/nvidia/hpc-benchmarks:23.10
 - dcgm diags
