UPDATE workflow_definitions
SET runtime_config = jsonb_set(
      runtime_config - 'analysis_model_before_vision_fix',
      '{analysis_model_code}',
      to_jsonb(runtime_config->>'analysis_model_before_vision_fix'),
      true
    ),
    updated_at = now()
WHERE code = 'ecommerce_image'
  AND runtime_config ? 'analysis_model_before_vision_fix';

UPDATE models
SET runtime_rule = runtime_rule #- '{capabilities,vision}',
    updated_at = now()
WHERE code = 'glm-4-6v';
