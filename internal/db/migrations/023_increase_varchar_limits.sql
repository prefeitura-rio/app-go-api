-- +goose Up
-- Increase VARCHAR limits to 20000 characters for text fields across all tables
-- Technical/enum fields (cpf, phone, status, modalidade, etc.) keep their original limits

-- CURSOS table
ALTER TABLE cursos ALTER COLUMN titulo TYPE VARCHAR(20000);
ALTER TABLE cursos ALTER COLUMN organization TYPE VARCHAR(20000);
ALTER TABLE cursos ALTER COLUMN theme TYPE VARCHAR(20000);
ALTER TABLE cursos ALTER COLUMN workload TYPE VARCHAR(20000);
ALTER TABLE cursos ALTER COLUMN target_audience TYPE VARCHAR(20000);
ALTER TABLE cursos ALTER COLUMN institutional_logo TYPE VARCHAR(20000);
ALTER TABLE cursos ALTER COLUMN cover_image TYPE VARCHAR(20000);
ALTER TABLE cursos ALTER COLUMN local_realizacao TYPE VARCHAR(20000);
ALTER TABLE cursos ALTER COLUMN link_inscricao TYPE VARCHAR(20000);
ALTER TABLE cursos ALTER COLUMN contato_duvidas TYPE VARCHAR(20000);
ALTER TABLE cursos ALTER COLUMN facilitator TYPE VARCHAR(20000);
ALTER TABLE cursos ALTER COLUMN external_partner_name TYPE VARCHAR(20000);
ALTER TABLE cursos ALTER COLUMN external_partner_url TYPE VARCHAR(20000);
ALTER TABLE cursos ALTER COLUMN external_partner_logo_url TYPE VARCHAR(20000);
ALTER TABLE cursos ALTER COLUMN external_partner_contact TYPE VARCHAR(20000);
ALTER TABLE cursos ALTER COLUMN accessibility TYPE VARCHAR(20000);

-- EMPREGOS table
ALTER TABLE empregos ALTER COLUMN titulo TYPE VARCHAR(20000);
ALTER TABLE empregos ALTER COLUMN contato_duvidas TYPE VARCHAR(20000);

-- INSCRICOES table
ALTER TABLE inscricoes ALTER COLUMN name TYPE VARCHAR(20000);
ALTER TABLE inscricoes ALTER COLUMN email TYPE VARCHAR(20000);
ALTER TABLE inscricoes ALTER COLUMN address TYPE VARCHAR(20000);
ALTER TABLE inscricoes ALTER COLUMN neighborhood TYPE VARCHAR(20000);
ALTER TABLE inscricoes ALTER COLUMN certificate_url TYPE VARCHAR(20000);

-- LOCATION_CLASSES table
ALTER TABLE location_classes ALTER COLUMN address TYPE VARCHAR(20000);
ALTER TABLE location_classes ALTER COLUMN neighborhood TYPE VARCHAR(20000);

-- COURSE_SCHEDULES table
ALTER TABLE course_schedules ALTER COLUMN class_time TYPE VARCHAR(20000);
ALTER TABLE course_schedules ALTER COLUMN class_days TYPE VARCHAR(20000);

-- REMOTE_CLASSES table
ALTER TABLE remote_classes ALTER COLUMN class_time TYPE VARCHAR(20000);
ALTER TABLE remote_classes ALTER COLUMN class_days TYPE VARCHAR(20000);

-- CUSTOM_FIELDS table
ALTER TABLE custom_fields ALTER COLUMN title TYPE VARCHAR(20000);

-- Related entities
ALTER TABLE categorias ALTER COLUMN nome TYPE VARCHAR(20000);
ALTER TABLE acessibilidades ALTER COLUMN nome TYPE VARCHAR(20000);
ALTER TABLE orgaos ALTER COLUMN nome TYPE VARCHAR(20000);
ALTER TABLE instituicoes_ensino ALTER COLUMN nome TYPE VARCHAR(20000);
ALTER TABLE empresas ALTER COLUMN nome TYPE VARCHAR(20000);
ALTER TABLE escolaridades ALTER COLUMN nivel TYPE VARCHAR(20000);

-- +goose Down
-- Rollback: restore original VARCHAR limits
-- Note: This may fail if data exceeds original limits

-- CURSOS table
ALTER TABLE cursos ALTER COLUMN titulo TYPE VARCHAR(255);
ALTER TABLE cursos ALTER COLUMN organization TYPE VARCHAR(255);
ALTER TABLE cursos ALTER COLUMN theme TYPE VARCHAR(100);
ALTER TABLE cursos ALTER COLUMN workload TYPE VARCHAR(50);
ALTER TABLE cursos ALTER COLUMN target_audience TYPE VARCHAR(600);
ALTER TABLE cursos ALTER COLUMN institutional_logo TYPE VARCHAR(500);
ALTER TABLE cursos ALTER COLUMN cover_image TYPE VARCHAR(500);
ALTER TABLE cursos ALTER COLUMN local_realizacao TYPE VARCHAR(255);
ALTER TABLE cursos ALTER COLUMN link_inscricao TYPE VARCHAR(255);
ALTER TABLE cursos ALTER COLUMN contato_duvidas TYPE VARCHAR(255);
ALTER TABLE cursos ALTER COLUMN facilitator TYPE VARCHAR(255);
ALTER TABLE cursos ALTER COLUMN external_partner_name TYPE VARCHAR(255);
ALTER TABLE cursos ALTER COLUMN external_partner_url TYPE VARCHAR(500);
ALTER TABLE cursos ALTER COLUMN external_partner_logo_url TYPE VARCHAR(500);
ALTER TABLE cursos ALTER COLUMN external_partner_contact TYPE VARCHAR(255);
ALTER TABLE cursos ALTER COLUMN accessibility TYPE VARCHAR(255);

-- EMPREGOS table
ALTER TABLE empregos ALTER COLUMN titulo TYPE VARCHAR(255);
ALTER TABLE empregos ALTER COLUMN contato_duvidas TYPE VARCHAR(255);

-- INSCRICOES table
ALTER TABLE inscricoes ALTER COLUMN name TYPE VARCHAR(255);
ALTER TABLE inscricoes ALTER COLUMN email TYPE VARCHAR(255);
ALTER TABLE inscricoes ALTER COLUMN address TYPE VARCHAR(500);
ALTER TABLE inscricoes ALTER COLUMN neighborhood TYPE VARCHAR(100);
ALTER TABLE inscricoes ALTER COLUMN certificate_url TYPE VARCHAR(500);

-- LOCATION_CLASSES table
ALTER TABLE location_classes ALTER COLUMN address TYPE VARCHAR(500);
ALTER TABLE location_classes ALTER COLUMN neighborhood TYPE VARCHAR(100);

-- COURSE_SCHEDULES table
ALTER TABLE course_schedules ALTER COLUMN class_time TYPE VARCHAR(50);
ALTER TABLE course_schedules ALTER COLUMN class_days TYPE VARCHAR(200);

-- REMOTE_CLASSES table
ALTER TABLE remote_classes ALTER COLUMN class_time TYPE VARCHAR(50);
ALTER TABLE remote_classes ALTER COLUMN class_days TYPE VARCHAR(200);

-- CUSTOM_FIELDS table
ALTER TABLE custom_fields ALTER COLUMN title TYPE VARCHAR(255);

-- Related entities
ALTER TABLE categorias ALTER COLUMN nome TYPE VARCHAR(100);
ALTER TABLE acessibilidades ALTER COLUMN nome TYPE VARCHAR(100);
ALTER TABLE orgaos ALTER COLUMN nome TYPE VARCHAR(255);
ALTER TABLE instituicoes_ensino ALTER COLUMN nome TYPE VARCHAR(255);
ALTER TABLE empresas ALTER COLUMN nome TYPE VARCHAR(255);
ALTER TABLE escolaridades ALTER COLUMN nivel TYPE VARCHAR(100);
