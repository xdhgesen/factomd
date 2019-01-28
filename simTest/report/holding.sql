\x

-- TODO: correlate w/ other metrics
SELECT
*
FROM (SELECT
  run,
  -- FIXME too hard to split
  --trim(BOTH '  ' FROM split_part((e->0)::varchar, ' ', 2)),
  e->0,
  *
FROM
  fnode0.holding) as h
LIMIT 10;
