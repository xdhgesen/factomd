

\x

/*
CREATE OR REPLACE FUNCTION ptnetState(varchar, varchar) RETURNS SETOF states
AS $$
DECLARE
BEGIN
  RETURN QUERY EXECUTE format('SELECT * FROM
    %I.states 
  WHERE
    ptnet = %L AND oid = %L
  ', $1, $1, $2);
  
END;
$$ LANGUAGE plpgsql;
*/


SELECT * from 
  (SELECT
    e->'ts' as ts,
    e->'seq' as seq,
    e->'log'->'height' as block, 
    e->'log'->'min' as min, 
    e->'log'->'event' as event,
    e->'log'->'event'->0 as e0,
    e->'log'->'event'->1 as e1,
    *
  FROM
    fnode0.simtest) as s
WHERE
  e0::text like '%HOLD%';


