# Performance Testing Audit

## Method:

* Boot up Prometheus w/ docker-compose
* Run TestLoadNewHolding 
  * to calibrate graphs looking at prometheus graphs
  * uses blktime=30s
* Loading single simulator node * gathering metrics from (I expect fnode0 ?)

### Test

* What a failure looks like
  * Every 5 min we count the change in block height 
  * blktime is currently set to 60s blocks,
    * We expect >=4 blocks during a 5 min span

```
--- FAIL: TestLoadNewHolding (1200.53s)
    LoadNewHolding_test.go:92: LLHT: 33<=>41 moved 8
    LoadNewHolding_test.go:86: only moved 2 blocks
```


Sample Pometheus graph

```
http://127.0.0.1:9090/graph?g0.range_input=15m&g0.stacked=1&g0.expr=rate(factomd_state_queue_total_general_inmsg_vec%7Bmessage%3D~%27(commitentry%7Crevealentry%7Cack).*%27%2C%20instance%3D%22factomd_2%3A9876%22%7D%5B60s%5D)&g0.tab=0&g1.range_input=15m&g1.stacked=1&g1.expr=rate(%7B__name__%20%3D~%20%27factomd_state_(total_receive)_time%27%2C%20instance%3D%22factomd_2%3A9876%22%20%7D%5B60s%5D)&g1.tab=0&g2.range_input=1h&g2.step_input=3&g2.stacked=0&g2.expr=rate(%7B__name__%20%3D~%20%27factomd_state_(ack_loop)_time%27%2C%20instance%3D%22factomd_2%3A9876%22%20%7D%5B60s%5D)%0A&g2.tab=0&g3.range_input=1h&g3.step_input=3&g3.stacked=1&g3.expr=rate(factomd_state_review_holding_time%7Binstance%3D%22factomd_2%3A9876%22%20%7D%5B60s%5D)&g3.tab=0&g4.range_input=1h&g4.stacked=1&g4.expr=rate(%7B__name__%20%3D~%20%27factomd_state_total_send_time%27%2C%20instance%3D%22factomd_2%3A9876%22%20%7D%5B60s%5D)&g4.tab=0
```
