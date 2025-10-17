## kubectl-ip-check

A kubectl plugin to check ip address resources across the cluster nodes. Provides overall allocated, used and free IPs in cluster.

Example: 
```
~ $ ./kubectl-ip-check-linux 
Allocated IPs: 682 | Used IPs: 170 | Free IPs: 512

NAME                              INSTANCE-ID           INSTANCE-TYPE   AVAILABILITY-ZONE   INTERNAL-IP       TOTAL-IPS   USED-IPS   FREE-IPS
ip-192-168-10-106.ec2.internal    i-064adaf03c2efb4f2   t3.small        us-east-1a          192.168.10.106    8           3          5
ip-192-168-101-147.ec2.internal   i-04f5d79942c711aa6   m5.2xlarge      us-east-1d          192.168.101.147   15          1          14
ip-192-168-103-152.ec2.internal   i-01ef79740a15eea5d   m5.2xlarge      us-east-1d          192.168.103.152   30          5          25
ip-192-168-11-45.ec2.internal     i-018fe96ee7b764915   t3.small        us-east-1a          192.168.11.45     8           3          5
ip-192-168-11-69.ec2.internal     i-00aef0ee2330568cb   t3.small        us-east-1a          192.168.11.69     8           3          5
ip-192-168-11-73.ec2.internal     i-06bdaf3673c7159b9   t3.small        us-east-1a          192.168.11.73     8           3          5
ip-192-168-111-254.ec2.internal   i-02781e32ca8d3539a   m5.2xlarge      us-east-1d          192.168.111.254   30          5          25
ip-192-168-116-45.ec2.internal    i-03fd3290d4503e43e   m5.2xlarge      us-east-1d          192.168.116.45    30          5          25
```
