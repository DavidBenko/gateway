ALTER TABLE proxy_endpoint_components
ADD COLUMN type_discriminator TEXT;

ALTER TABLE proxy_endpoint_components
ADD COLUMN name TEXT;

ALTER TABLE proxy_endpoint_components
ADD COLUMN description TEXT;

ALTER TABLE proxy_endpoint_components
ADD COLUMN api_id INTEGER;

UPDATE proxy_endpoint_components AS pec
SET api_id = pe.api_id, type_discriminator = 'standard'
FROM proxy_endpoints pe
WHERE pe.id = pec.endpoint_id;

ALTER TABLE proxy_endpoint_components
ADD FOREIGN KEY("api_id") REFERENCES "apis"("id") ON DELETE CASCADE;

ALTER TABLE proxy_endpoint_components
ADD UNIQUE("api_id", "name");

CREATE TABLE IF NOT EXISTS "proxy_endpoint_component_references" (
  "id" SERIAL PRIMARY KEY,
  "proxy_endpoint_id" INTEGER NOT NULL,
  "position" INTEGER NOT NULL,
  "proxy_endpoint_component_id" INTEGER NOT NULL,
  FOREIGN KEY("proxy_endpoint_id") REFERENCES "proxy_endpoints"("id") ON DELETE CASCADE,
  FOREIGN KEY("proxy_endpoint_component_id") REFERENCES "proxy_endpoint_components"("id") ON DELETE NO ACTION
);

INSERT INTO proxy_endpoint_component_references(proxy_endpoint_id, position, proxy_endpoint_component_id)
SELECT endpoint_id, position, id as proxy_endpoint_component_id
FROM proxy_endpoint_components;

CREATE OR REPLACE FUNCTION proxy_endpoint_component_references_conditional_delete() RETURNS TRIGGER AS $plpgsql$
  DECLARE
    num_refs integer;
  BEGIN
    SELECT COUNT(1) INTO num_refs
    FROM proxy_endpoint_component_references
    WHERE proxy_endpoint_component_id = OLD.proxy_endpoint_component_id;

    IF num_refs = 0 AND TG_OP = 'DELETE' THEN
      DELETE FROM proxy_endpoint_components
      WHERE type_discriminator = 'standard'
      AND id = OLD.proxy_endpoint_component_id;
    END IF;

    RETURN NULL;
  END;
$plpgsql$ LANGUAGE plpgsql;

CREATE TRIGGER proxy_endpoint_component_references_conditional_delete_trigger
AFTER DELETE ON proxy_endpoint_component_references
FOR EACH ROW EXECUTE PROCEDURE proxy_endpoint_component_references_conditional_delete();

ALTER TABLE proxy_endpoint_components
DROP COLUMN endpoint_id;

ALTER TABLE proxy_endpoint_components
DROP COLUMN position;

CREATE INDEX idx_proxy_endpoint_components_type_discriminator ON proxy_endpoint_components USING btree(type_discriminator);
CREATE INDEX idx_proxy_endpoint_components_api_id ON proxy_endpoint_components USING btree(api_id);
CREATE INDEX idx_proxy_endpoint_component_references_proxy_endpoint_id ON proxy_endpoint_component_references USING btree(proxy_endpoint_id);
CREATE INDEX idx_proxy_endpoint_component_references_proxy_endpoint_component_id ON proxy_endpoint_component_references USING btree(proxy_endpoint_component_id);

ANALYZE;
