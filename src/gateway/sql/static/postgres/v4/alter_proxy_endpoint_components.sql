ALTER TABLE proxy_endpoint_components
  ADD COLUMN shared_component_id INTEGER,
  ADD FOREIGN KEY("shared_component_id")
    REFERENCES "shared_components"("id")
    ON DELETE SET NULL;
