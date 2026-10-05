-- Number 041 reserves 029-040 for the separately reviewed operational branch.
CREATE TABLE attendance_request (
 actor_id text NOT NULL REFERENCES "user"(id),
 operation text NOT NULL,
 request_key text NOT NULL CHECK(length(request_key) BETWEEN 1 AND 200),
 input jsonb NOT NULL,
 response jsonb NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(actor_id, operation, request_key)
);
