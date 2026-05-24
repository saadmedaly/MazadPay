-- Migration to remove messaging/chat features
DROP VIEW IF EXISTS user_conversations;
DROP TABLE IF EXISTS message_status;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS conversation_participants;
DROP TABLE IF EXISTS conversations;
