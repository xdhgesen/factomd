# Performance Testing Audit


## Method:

* Boot up Prometheus w/ docker-compose
* Run TestLoadNewHolding 
  * to calibrate graphs looking at prometheus graphs
  * uses blktime=30s
* Loading single simulator node * gathering metrics from (I expect fnode0 ?)


### TPS

Counting the rate of revealentry messages seems to correlate w/ the overall TPS

```
rate(factomd_state_queue_total_general_inmsg_vec{message=~"(commitentry|revealentry|ack).*"}[30s])/30
```

### P2P/Networking

It seems like we don't get any p2p metrics during sim testing
every `factomd_p2p_*` is 0 during simtesting

? how can we measure p2p traffic?


These correlate w/ TPS
```
rate({__name__ =~ 'factomd_state_total_send_time'}[30s])
rate({__name__ =~ 'factomd_state_total_receive_time'}[30s])
```


### Backlog/holding

Holding status don't seem to work

```
{__name__ =~ '.*holding.*'}
```

Holding review time Does correlate w/ TPS graphs

```
rate(factomd_state_review_holding_time{}[30s])
```
