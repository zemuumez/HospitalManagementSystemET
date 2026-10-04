CREATE TRIGGER immutable_audit_event BEFORE UPDATE OR DELETE ON audit_event
 FOR EACH ROW EXECUTE FUNCTION protect_pharmacy_history();
CREATE TRIGGER no_truncate_audit_event BEFORE TRUNCATE ON audit_event
 FOR EACH STATEMENT EXECUTE FUNCTION protect_pharmacy_history();
CREATE INDEX audit_event_actor_cursor ON audit_event(actor_id,id DESC);
CREATE INDEX audit_event_action_cursor ON audit_event(action,id DESC);
CREATE INDEX audit_event_resource_cursor ON audit_event(resource_id,id DESC);
