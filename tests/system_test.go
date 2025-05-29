package tests

import (
	"math/rand"
	"testing"
	"time"

	"github.com/IlianBuh/Post-service/internal/config"
	"github.com/IlianBuh/Post-service/tests/suite"
	userinfov1 "github.com/IlianBuh/SSO_Protobuf/gen/go/userinfo"
	"github.com/brianvoe/gofakeit"
	"github.com/stretchr/testify/require"
)

func TestSystem(t *testing.T) {
	cfg := config.MustLoad("/home/il/Pet-project/Post-Service/config/config.json")
	s := suite.NewSuite(t, cfg)
	u, err := suite.NewUserClient("localhost:20202")
	require.NoError(t, err)

	for i := 0; i < 2000; i++ {
		user, err := u.Client.User(
			t.Context(),
			&userinfov1.UserRequest{
				Uuid: int32(rand.Uint32()%200 + 1),
			},
		)
		require.NoError(t, err)

		s.Post.Create(
			t.Context(),
			int(user.GetUser().Uuid),
			user.User.Login,
			gofakeit.Sentence(int((rand.Uint32()%20)+5)),
			gofakeit.Paragraph(int(rand.Uint32()%2+1), 3, int((rand.Uint32()%25)+10), " "),
			generateThemes(),
		)
	}

	time.Sleep(time.Second * 10)
}

func generateThemes() []string {
	allThemes := []string{
		"tech", "life", "coding", "travel", "food", "fitness", "education",
		"art", "music", "books", "startup", "finance", "career", "design",
	}

	num := rand.Intn(3) + 1 // 1 to 3 themes per post
	themes := make(map[string]bool)
	var selected []string
	for len(selected) < num {
		theme := allThemes[rand.Intn(len(allThemes))]
		if !themes[theme] {
			themes[theme] = true
			selected = append(selected, theme)
		}
	}
	return selected
}
