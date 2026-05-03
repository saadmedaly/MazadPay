-- Migration to revert user_conversations view (removing updated_at)

DROP VIEW IF EXISTS user_conversations;

CREATE VIEW user_conversations AS
SELECT 
    cp.user_id,
    c.id as conversation_id,
    c.type,
    c.title,
    c.last_message_at,
    c.last_message_preview,
    c.last_message_sender_id,
    c.is_active,
    cp.role,
    cp.joined_at,
    cp.last_read_at,
    cp.unread_count,
    cp.is_muted
FROM conversation_participants cp
JOIN conversations c ON cp.conversation_id = c.id
WHERE c.is_active = TRUE;
