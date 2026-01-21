-- +goose Up
CREATE TABLE customers (
                         user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,

  -- The Stripe Customer ID (e.g., 'cus_xxxxxxxxxxxxxx')
                         stripe_customer_id VARCHAR(255) UNIQUE NOT NULL,

  -- The Stripe ID of the user's default payment method.
                         default_payment_method_id VARCHAR(255),

                         created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                         updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TYPE subscription_tier AS ENUM ('free', 'premium', 'pro');
CREATE TABLE subscriptions (
                             id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                             user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- Assuming one subscription per user

                             tier subscription_tier NOT NULL,
                             status subscription_status NOT NULL,

  -- Stripe IDs are the link to the source of truth
                             stripe_subscription_id VARCHAR(255) UNIQUE NOT NULL,
                             stripe_price_id VARCHAR(255) NOT NULL,

  -- Cache important data from Stripe for fast lookups
                             current_period_start TIMESTAMPTZ,
                             current_period_end TIMESTAMPTZ,
                             cancel_at_period_end BOOLEAN NOT NULL DEFAULT false,
                             cancelled_at TIMESTAMPTZ,

                             created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                             updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TYPE payment_purpose AS ENUM ('workshop', 'gig', 'certification', 'premium_chat');
CREATE TYPE payment_status AS ENUM ('requires_action', 'pending', 'succeeded', 'failed', 'refunded');

CREATE TABLE payments (
                        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                        user_id UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,

  -- Stripe PaymentIntent ID is the source of truth
                        stripe_payment_intent_id VARCHAR(255) UNIQUE NOT NULL,

                        amount NUMERIC(10, 2) NOT NULL,
                        currency VARCHAR(3) NOT NULL,

                        purpose payment_purpose NOT NULL,
                        status payment_status NOT NULL,
                        description TEXT,

  -- Store related entity IDs in a flexible way
                        metadata JSONB, -- e.g., {"gig_id": "...", "workshop_id": "..."}

                        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payments_user_id ON payments (user_id);
CREATE INDEX idx_payments_stripe_payment_intent_id ON payments (stripe_payment_intent_id);

CREATE TYPE escrow_status AS ENUM ('holding', 'released', 'refunded', 'disputed');

CREATE TABLE escrow_payments (
                               id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                               gig_id UUID NOT NULL REFERENCES gigs(id), -- Assuming a 'gigs' table

                               payer_id UUID NOT NULL REFERENCES users(id),
                               payee_id UUID NOT NULL REFERENCES users(id),

  -- Corresponds to a Stripe Transfer or similar object
                               stripe_transfer_id VARCHAR(255),

                               amount NUMERIC(10, 2) NOT NULL,
                               platform_fee NUMERIC(10, 2) NOT NULL,

                               status escrow_status NOT NULL DEFAULT 'holding',

                               created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                               release_at TIMESTAMPTZ, -- The scheduled auto-release time
                               released_at TIMESTAMPTZ -- When it was actually released
);

CREATE INDEX idx_escrow_payments_gig_id ON escrow_payments (gig_id);

CREATE TYPE payout_status AS ENUM ('pending', 'in_transit', 'paid', 'failed', 'canceled');

CREATE TABLE payouts (
                       id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                       user_id UUID NOT NULL REFERENCES users(id), -- The user receiving the payout

  -- Stripe Payout ID is the source of truth
                       stripe_payout_id VARCHAR(255) UNIQUE NOT NULL,

                       amount NUMERIC(10, 2) NOT NULL,
                       currency VARCHAR(3) NOT NULL,

                       status payout_status NOT NULL DEFAULT 'pending',

  -- Stripe Connect destination account ID
                       destination_account_id VARCHAR(255) NOT NULL,

                       initiated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                       estimated_arrival_at TIMESTAMPTZ
);

CREATE INDEX idx_payouts_user_id ON payouts (user_id);

CREATE TABLE invoices (
                        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                        user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                        subscription_id UUID REFERENCES subscriptions(id),

  -- Stripe Invoice ID is the source of truth
                        stripe_invoice_id VARCHAR(255) UNIQUE NOT NULL,

                        amount_due NUMERIC(10, 2) NOT NULL,
                        amount_paid NUMERIC(10, 2) NOT NULL,

                        status VARCHAR(50) NOT NULL, -- e.g., 'draft', 'open', 'paid', 'uncollectible'

                        due_date TIMESTAMPTZ,
                        paid_at TIMESTAMPTZ,

  -- The URL to the PDF hosted by Stripe
                        invoice_pdf_url TEXT,

                        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

