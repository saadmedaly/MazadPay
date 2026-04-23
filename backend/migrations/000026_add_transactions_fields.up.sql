-- ============================================================
-- Ajout de champs manquants à la table transactions
-- Gestion reçus, méthode de paiement, frais, détails
-- ============================================================

-- Ajouter receipt_image_temp (image uploadée avant validation)
ALTER TABLE transactions ADD COLUMN receipt_image_temp TEXT;

-- Ajouter payment_method (méthode spécifique)
ALTER TABLE transactions ADD COLUMN payment_method VARCHAR(50);
CREATE INDEX idx_transactions_payment_method ON transactions(payment_method);

-- Ajouter fee_amount (montant des frais)
ALTER TABLE transactions ADD COLUMN fee_amount DECIMAL(15,2) DEFAULT 0.00;

-- Ajouter net_amount (montant net après frais)
ALTER TABLE transactions ADD COLUMN net_amount DECIMAL(15,2);

-- Ajouter description (description personnalisée)
ALTER TABLE transactions ADD COLUMN description TEXT;

-- Ajouter failure_reason (raison de l'échec)
ALTER TABLE transactions ADD COLUMN failure_reason TEXT;
