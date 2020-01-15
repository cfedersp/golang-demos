		Rate-Limited Priority Queue
Overview:
Given 2 queues, a main queue and a priority queue:
Immediately process the main queue, then when items are available on the priority queue, only process the priority queue until it is empty.
Integrate with a throttled service at a suitably low rate that still supports concurrent workers.

Run:
go run int-framework.go jr-candidates.txt sr-candidates.txt

