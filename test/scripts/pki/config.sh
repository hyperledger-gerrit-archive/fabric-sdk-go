#!/usr/bin/env bash

# SERVERS[CN]=HOSTNAME1,HOSTNAME2,...
declare -A SERVERS=(
  [ca.org1.example.com]="ca.org1.example.com"
  [ca.org2.example.com]="ca.org2.example.com"
  [peer0.org1.example.com]="peer0.org1.example.com"
  [peer1.org1.example.com]="peer1.org1.example.com"
  [peer0.org2.example.com]="peer0.org2.example.com"
  [peer1.org2.example.com]="peer1.org2.example.com"
  [orderer.example.com]="orderer.example.com"
  [wild.org1.example.com]="*.org1.example.com"
  [wild.org2.example.com]="*.org2.example.com"
  [wild_both_orgs]="*.org1.example.com,*.org2.example.com"
)

# CLIENTS[CN]=HOSTNAMES
declare -A CLIENTS=(
  [fabric-client]="fabric-client"
)
