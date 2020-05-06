# dns-perf-mon

This DNS performance monitor tool was build in order to keep track of the DNS performance from the [Pi-hole](https://pi-hole.net/) that I installed in my home network. Since this Pi-hole is currently running on a Raspberry Pi Model B (Gen 1), which is not very powerful, I needed a tool to track this.

The `dns-perf-mon` runs daemonized and loads a list of domains from the `domains.txt` file (customize it whatever you want, 1 DNS domain per line) and stores all lookup information in the `dns-query-times.csv` file.

A most powerful analysis can be made using [pandas](https://pandas.pydata.org/) with this output dataset.

As soon as I have time for it I'll improve the current implementation, which is currently very rudimentary.