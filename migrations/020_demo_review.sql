WITH standard_seed(id, version, status, revision, request_key, document) AS (
    VALUES (
        'quality-standard'::text,
        1,
        'published'::text,
        1::bigint,
        'seed-template'::text,
        jsonb_build_object(
            'ID','quality-standard','DomainID','production','Name','生产质量评议','Version',1,'Status','published','Revision',1,
            'Criteria',jsonb_build_array(
                jsonb_build_object('ID','quality','Title','质量评分','Required',true,'Weight',1,'Scale',jsonb_build_object('Min',0,'Max',5,'Pass',3)),
                jsonb_build_object('ID','safety-veto','Title','安全否决','Required',true,'Veto',true,'Weight',0,'Scale',jsonb_build_object('Min',0,'Max',1,'Pass',1))
            )
        )
    )
)
INSERT INTO templates(payload, idempotency_key, revision, status, version, id)
SELECT document, request_key, revision, status, version, id FROM standard_seed
ON CONFLICT DO NOTHING;

WITH batch_seed AS (
    SELECT
        'batch-1'::text AS id,
        'submitted'::text AS status,
        2::bigint AS revision,
        'seed-batch'::text AS request_key,
        jsonb_build_object(
            'ID','batch-1','ProjectID','demo-project','Name','新品首件评议','Status','submitted','Revision',2,
            'Snapshot',jsonb_build_object('TemplateID','quality-standard','Version',1,'Name','生产质量评议','Criteria',jsonb_build_array(
                jsonb_build_object('ID','quality','Title','质量评分','Required',true,'Weight',1,'Scale',jsonb_build_object('Max',5)),
                jsonb_build_object('ID','safety-veto','Title','安全否决','Required',true,'Veto',true,'Scale',jsonb_build_object('Max',1))
            )),
            'Materials',jsonb_build_array(jsonb_build_object('ID','material-1','Title','新能源壳体首件')),
            'Assignments',jsonb_build_array(jsonb_build_object('ReviewerID','reviewer-a','MaterialID','material-1'),jsonb_build_object('ReviewerID','reviewer-b','MaterialID','material-1'))
        ) AS document
)
INSERT INTO batches(payload, template_version, template_id, idempotency_key, revision, status, id)
SELECT document, 1, 'quality-standard', request_key, revision, status, id FROM batch_seed
ON CONFLICT DO NOTHING;
