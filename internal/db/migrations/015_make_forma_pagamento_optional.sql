-- +goose Up
-- +goose StatementBegin

-- Tornar o campo forma_pagamento opcional (nullable) na tabela oportunidades_mei
ALTER TABLE oportunidades_mei
ALTER COLUMN forma_pagamento DROP NOT NULL;

-- Adicionar comentário explicando os valores aceitos
COMMENT ON COLUMN oportunidades_mei.forma_pagamento IS 'Forma de pagamento. Valores aceitos: CHEQUE, DINHEIRO, CARTAO, PIX, TRANSFERENCIA ou NULL/vazio (campo opcional)';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Reverter: tornar o campo forma_pagamento obrigatório novamente
-- Nota: Esta operação define PIX como padrão para registros NULL antes de tornar o campo obrigatório

-- Primeiro, definir um valor padrão para registros com NULL ou vazio
UPDATE oportunidades_mei
SET forma_pagamento = 'PIX'
WHERE forma_pagamento IS NULL OR forma_pagamento = '';

-- Depois tornar o campo NOT NULL novamente
ALTER TABLE oportunidades_mei
ALTER COLUMN forma_pagamento SET NOT NULL;

-- Remover o comentário
COMMENT ON COLUMN oportunidades_mei.forma_pagamento IS NULL;

-- +goose StatementEnd