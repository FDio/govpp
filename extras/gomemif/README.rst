.. _gomemif_doc:

Gomemif library
=======================

Memif library implemented in Go. The package contains 3 examples: Bridge and ICMP responder in interrupt and polling mode.

setup and run
-------------
To Build all examples

::

   cd .. && make extras

To Run ICMP responder in interrupt mode:

::

   DBGvpp# create interface memif id 0 master
   DBGvpp# set int ip addr memif0/0 192.168.1.2/24
   DBGvpp# set int state memif0/0 up

   examples/icmp_responder_cb/icmp_responder_cb
   gomemif# start

   DBGvpp# ping 192.168.1.1

To Run ICMP responder in polling mode:

::

   DBGvpp# create interface memif id 0 master
   DBGvpp# set int ip addr memif0/0 192.168.1.2/24
   DBGvpp# set int state memif0/0 up

   examples/icmp_responder_poll/icmp_responder_poll
   gomemif# start

   DBGvpp# ping 192.168.1.1

To Run Bridge:

::

  examples/bridge/bridge
  gomemif# start



