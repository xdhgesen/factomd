

-- query run times
SELECT
  run,
  trim('"' FROM seq::text )::int as seq,
  trim('"' FROM block::text )::int as block,
  trim('"' FROM min::text )::int as min,
  --e0,
  e1::decimal/100000000 as batch_sec, -- time in sec as 
  --e2
  ts::text::time

FROM 
  (SELECT
    run,
    e->'ts' as ts,
    e->'seq' as seq,
    e->'log'->'height' as block, 
    e->'log'->'min' as min, 
    --e->'log'->'event' as event,
    e->'log'->'event'->0 as e0,
    e->'log'->'event'->1 as e1,
    e->'log'->'event'->2 as e2
  FROM
    fnode0.simtest) as s
WHERE
  e2::text like '%RUNTIME%'
ORDER BY
  run, seq ;
