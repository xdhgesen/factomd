
-- query size of holding (and all queues/ lists)

SELECT
  trim('"' FROM block::text )::int as block,
  trim('"' FROM min::text )::int as min,
  s.holding,
  s.acks,
  s.msgqueue,
  s.inmsgqueue,
  s.apiqueue,
  s.ackqueue,
  s.timermsg,
  ts::text::time
FROM (SELECT
  --mt     | "LIST_SIZES Holding: %v, Acks: %v, MsgQueue: %v, InMsgQueue: %v, APIQueue: %v, AckQueue: %v, TimerMsg: %v "
  run,
  e->'log'->'height' as block,
  e->'log'->'min' as min,
  e->'log'->'event'->0 as holding,
  e->'log'->'event'->1 as acks,
  e->'log'->'event'->2 as msgqueue,
  e->'log'->'event'->3 as inmsgqueue,
  e->'log'->'event'->4 as apiqueue,
  e->'log'->'event'->5 as ackqueue,
  e->'log'->'event'->6 as timermsg,
  e->'log'->'event'->7 as fmt,
  e->'ts' as ts
  --e as l
FROM
  logs
) as s

WHERE
  s.fmt::varchar like '%LIST_SIZES%'
ORDER BY ts;
