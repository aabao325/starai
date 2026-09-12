-- Ecommerce analysis receives product reference images, so it must use a model
-- that accepts image content. Preserve the previous binding for a clean rollback.
UPDATE models
SET runtime_rule = jsonb_set(runtime_rule, '{capabilities,vision}', 'true'::jsonb, true),
    updated_at = now()
WHERE code = 'glm-4-6v';

UPDATE workflow_definitions
SET runtime_config = jsonb_set(
      jsonb_set(runtime_config, '{analysis_model_before_vision_fix}', to_jsonb(COALESCE(runtime_config->>'analysis_model_code', '')), true),
      '{analysis_model_code}',
      '"glm-4-6v"'::jsonb,
      true
    ),
    updated_at = now()
WHERE code = 'ecommerce_image'
  AND COALESCE(runtime_config->>'analysis_model_code', '') IN ('', 'glm-4-7-flash', 'glm-5-3-flash')
  AND EXISTS (SELECT 1 FROM models WHERE code = 'glm-4-6v' AND is_enabled = true AND category = 'chat');
