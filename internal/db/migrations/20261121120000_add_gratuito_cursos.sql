-- +goose Up
-- DEV-715: o curso passa a informar se é gratuito (true) ou pago (false), para
-- o filtro gratuito da busca do app-catalogo. Os cursos existentes ficam
-- gratuitos, o padrão da coluna (os cursos da Prefeitura são oferecidos
-- gratuitamente); o curso pago é marcado com gratuito=false no cadastro.
-- Só acrescenta (expand): o código anterior não lê a coluna e o DEFAULT cobre
-- os INSERTs dele. Com DEFAULT constante, o ADD COLUMN não reescreve a tabela.
-- IF NOT EXISTS: o AutoMigrate do GORM, que roda depois, não falha nem repete.
ALTER TABLE cursos ADD COLUMN IF NOT EXISTS gratuito BOOLEAN NOT NULL DEFAULT true;

-- +goose Down
ALTER TABLE cursos DROP COLUMN IF EXISTS gratuito;
