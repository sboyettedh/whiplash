# whiplash
An instrumentation system for Ceph clusters.

Whiplash tells you things about the state of your Ceph cluster. It
doesn't tell you anything that Ceph doesn't already tell you, but it
does a lot more aggregating and filtering and massaging of
data. There'll be more to say about this after there's some working
code.

It also provides an out-of-band mechanism for ascertaining the state
of a cluster, which can be invaluable when the cluster itself is under
stress or is behaving anomalously.

Whiplash is written in Go, and so it is named after the family
Mastigoteuthidae (Masti-go-teuth-idae), the whip-lash squid. You're
welcome.