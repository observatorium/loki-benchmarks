# Benchmark Report

## Goal
We want to give the user the opportunity to know the recommended LokiStack configuration of OpenShift cluster that can stand for his daily logs written and read from Loki. Our goal is to create a table that specifies for each "T-shirt" size the average and max daily read and write capabilities of Loki.

## Environment
We ran the benchmark on OpenShift running on AWS clusters with configuration:

	compute:
	- architecture: amd64
	hyperthreading: Enabled
	name: worker
	platform:
		aws:
		type: m4.16xlarge
		rootVolume:
			type: io1
			size: 500
			iops: 8000
	replicas: 6
	controlPlane:
	architecture: amd64
	hyperthreading: Enabled
	name: master
	platform: {}
	replicas: 3

In this configuration each node has 64 vCPUs and 256 Memory (GiB)


## The Proccess Of Creating The Table
The benchmark tool records operational metrics (such as CPU, Memory, etc...) of Loki while sending a different amount of logs in each test, the logs are sent by a logger that is deployed when the benchmark test is started (the implementation can be seen at https://github.com/ViaQ/cluster-logging-load-client).

We have been observing mainly the CPU (in MiliCore) and Memory (in GB) as a function of the GiPD (Log load in GB per Day) to see how they change. We ran the benchmark test with different values of GiPD and watched over the CPU and Memory and how the number of Ingesters may affect these values in Loki. First, we ran Loki with 2 Ingesters and then we ran it with 4 Ingesters.

Then a high amount of logs has been sent to Loki by running different values between 2250GB and 6000GB per day (starting with 2250Gb then 2500GB ... and so on until 6000GB).


## The Commands That Have Been Used:

Observing the GiPD:

	`sum(rate(loki_distributor_bytes_received_total{[label]=~".*[job].*"}[duration])) / BytesToGigabytesMultiplier * 86400` 
	we devide by BytesToGigabytesMultiplier * 86400 in order to convert to Gi per day

Getting the CPU:

	`sum(rate(container_cpu_user_seconds_total{[label]=~".*[job].*"[duration]]) * 1000)`

Getting the Memory:

	`sum(avg_over_time(container_memory_working_set_bytes{[label]=~".*[job].*"}[duration]) / %BytesToGigabytesMultiplier)`
	we devide to convert to Gi


## Graphs

<img src="./low-cpu.png" alt="low load cpu" width="600"/>

<img src="./low-memory.png" alt="low load memory" width="600"/>

<img src="./high-cpu.png" alt="high load cpu" width="600"/>

<img src="./high-memory.png" alt="high load memory" width="600"/>


## Conclusions:
1. From the results we can see:

	 2 Ingesters: 

		when the GiPD is 2000 the Memory got above what is requested, and when the GiPD is about 3000 the Memory reaches it's limit and CPU reaches what is requested.

	 4 Ingesters: 
		
		when the GiPD is 2000 75% of the requested Memory is used (which is 37.5% of all the Memory) and 37.5% of the requested CPU is used (which is 18.75% of all the CPU).

2. It's clear that Loki with 2 Ingesters is not capable to bear 2000GiPD and above while 4 Ingesters can bear this amount of GiPD. However, what is not clear about this information is that sending a high amount of GB to Loki with 2 Ingesters made the GiPD go too high compared to what has been sent, but the GiPD in Loki with 4 Ingesters was pretty much normal.


