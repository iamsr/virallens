package testutil_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/yourusername/virallens/backend/internal/domain"
	"github.com/yourusername/virallens/backend/test/testutil"
)

func TestCreateTestUser(t *testing.T) {
	user := testutil.CreateTestUser()

	testutil.AssertValidUUID(t, user.ID, "User ID should be valid")
	if user.Username == "" {
		t.Error("Username should not be empty")
	}
	if user.Email == "" {
		t.Error("Email should not be empty")
	}
	if user.PasswordHash == "" {
		t.Error("PasswordHash should not be empty")
	}
	testutil.AssertTimeRecent(t, user.CreatedAt, "CreatedAt should be recent")
	testutil.AssertTimeRecent(t, user.UpdatedAt, "UpdatedAt should be recent")
}

func TestCreateTestConversation(t *testing.T) {
	t.Run("empty conversation", func(t *testing.T) {
		conv := testutil.CreateTestConversation()

		testutil.AssertValidUUID(t, conv.ID, "Conversation ID should be valid")
		if len(conv.Participants) != 0 {
			t.Errorf("Expected 0 participants, got %d", len(conv.Participants))
		}
		testutil.AssertTimeRecent(t, conv.CreatedAt, "CreatedAt should be recent")
		testutil.AssertTimeRecent(t, conv.UpdatedAt, "UpdatedAt should be recent")
	})

	t.Run("conversation with participants", func(t *testing.T) {
		user1ID := uuid.New()
		user2ID := uuid.New()
		conv := testutil.CreateTestConversation(user1ID, user2ID)

		testutil.AssertValidUUID(t, conv.ID, "Conversation ID should be valid")
		if len(conv.Participants) != 2 {
			t.Errorf("Expected 2 participants, got %d", len(conv.Participants))
		}
		if conv.Participants[0] != user1ID {
			t.Errorf("First participant should be %v, got %v", user1ID, conv.Participants[0])
		}
		if conv.Participants[1] != user2ID {
			t.Errorf("Second participant should be %v, got %v", user2ID, conv.Participants[1])
		}
		testutil.AssertTimeRecent(t, conv.CreatedAt, "CreatedAt should be recent")
		testutil.AssertTimeRecent(t, conv.UpdatedAt, "UpdatedAt should be recent")
	})
}

func TestCreateTestGroup(t *testing.T) {
	creatorID := uuid.New()
	group := testutil.CreateTestGroup(creatorID)

	testutil.AssertValidUUID(t, group.ID, "Group ID should be valid")
	if group.Name == "" {
		t.Error("Group name should not be empty")
	}
	if group.CreatedBy != creatorID {
		t.Errorf("CreatedBy should be %v, got %v", creatorID, group.CreatedBy)
	}
	testutil.AssertTimeRecent(t, group.CreatedAt, "CreatedAt should be recent")
	testutil.AssertTimeRecent(t, group.UpdatedAt, "UpdatedAt should be recent")
}

func TestCreateTestMessage(t *testing.T) {
	senderID := uuid.New()
	conversationID := uuid.New()
	msg := testutil.CreateTestMessage(senderID, conversationID)

	testutil.AssertValidUUID(t, msg.ID, "Message ID should be valid")
	if msg.SenderID != senderID {
		t.Errorf("SenderID should be %v, got %v", senderID, msg.SenderID)
	}
	if msg.ConversationID == nil || *msg.ConversationID != conversationID {
		t.Errorf("ConversationID should be %v, got %v", conversationID, msg.ConversationID)
	}
	if msg.GroupID != nil {
		t.Error("GroupID should be nil for conversation messages")
	}
	if msg.Content == "" {
		t.Error("Content should not be empty")
	}
	if msg.Type != domain.MessageTypeConversation {
		t.Errorf("Type should be %v, got %v", domain.MessageTypeConversation, msg.Type)
	}
	testutil.AssertTimeRecent(t, msg.CreatedAt, "CreatedAt should be recent")
}

func TestAssertValidUUID(t *testing.T) {
	// This would pass
	validID := uuid.New()
	testutil.AssertValidUUID(t, validID, "Should accept valid UUID")
}

func TestAssertNoError(t *testing.T) {
	// This would pass
	testutil.AssertNoError(t, nil, "test operation")
}

func TestAssertError(t *testing.T) {
	// Verify the function is accessible by calling it with a known error
	// This is just a compilation check - the function exists and is usable
	_ = testutil.AssertError
}
