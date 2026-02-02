-- +goose Up
CREATE TABLE points_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    points_change INT NOT NULL,
    action VARCHAR(255) NOT NULL,
    new_total INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_points_history_user_id ON points_history (user_id);
CREATE INDEX idx_points_history_action ON points_history (action);


-- +goose Down
DROP TABLE IF EXISTS points_history;
