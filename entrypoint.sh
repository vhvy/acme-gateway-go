#!/bin/sh

GATEWAY_IP="127.0.0.1"
GATEWAY_PORT="5353"
echo "Redirecting all DNS traffic to ${GATEWAY_IP}:${GATEWAY_PORT}"
iptables -t nat -A OUTPUT -p udp --dport 53 -j DNAT --to-destination ${GATEWAY_IP}:${GATEWAY_PORT}
iptables -t nat -A OUTPUT -p tcp --dport 53 -j DNAT --to-destination ${GATEWAY_IP}:${GATEWAY_PORT}

exec "$@"