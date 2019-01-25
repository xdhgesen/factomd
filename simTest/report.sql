

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


--SELECT * FROM fnode0.simtest limit 1;

CREATE OR REPLACE FUNCTION simtest_report() RETURNS SETOF RECORD as $$
  SELECT
    e->'ts' as ts,
    e->'seq' as seq,
    e->'log'->'height' as block, 
    e->'log'->'min' as min, 
    e->'log'->'event' as event, *
  FROM
    fnode0.simtest; 
$$ LANGUAGE sql;

SELECT simtest_report()

