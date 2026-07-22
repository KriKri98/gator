package cli

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/KriKri98/gator/internal/config"
	"github.com/KriKri98/gator/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Status struct {
	Cfg *config.Config
	DB  *database.Queries
}

type Command struct {
	Name string
	Args []string
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

type Commands struct {
	Command map[string]func(*Status, Command) error
}

func (c *Commands) Run(s *Status, cmd Command) error {
	err := c.Command[cmd.Name](s, cmd)
	if err != nil {
		return err
	}
	return nil
}

func (c *Commands) Register(name string, f func(*Status, Command) error) {
	c.Command[name] = f
}

func MiddlewareLoggedIn(handler func(s *Status, cmd Command, user database.User) error) func(s *Status, cmd Command) error {
	return func(s *Status, cmd Command) error {
		user, err := s.DB.GetUser(context.Background(), s.Cfg.Current_user_name)
		if err != nil {
			return err
		}

		return handler(s, cmd, user)
	}

}

func HandlerLogin(s *Status, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("no username given")
	}

	user, err := s.DB.GetUser(context.Background(), cmd.Args[0])
	if err != nil {
		fmt.Printf("username does not exist")
		os.Exit(1)
	}

	err = s.Cfg.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Printf("User %v has been set\n", user.Name)
	return nil
}

func HandlerRegister(s *Status, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("no username given")
	}
	userParams := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.Args[0],
	}
	user, err := s.DB.CreateUser(context.Background(), userParams)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	err = s.Cfg.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Printf("created user: %v", user)

	return nil
}

func HandlerReset(s *Status, cmd Command) error {
	err := s.DB.DeleteUsers(context.Background())
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
	return nil
}

func HandlerGetUsers(s *Status, cmd Command) error {
	users, err := s.DB.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range users {
		if user.Name == s.Cfg.Current_user_name {
			fmt.Printf("* %v (current)\n", user.Name)
		} else {
			fmt.Printf("* %v\n", user.Name)
		}
	}
	return nil
}

func fetchFeed(ctx context.Context, feedUrl string) (*RSSFeed, error) {
	feed := &RSSFeed{}
	req, err := http.NewRequestWithContext(ctx, "GET", feedUrl, http.NoBody)
	if err != nil {
		return feed, err
	}

	req.Header.Set("User-Agent", "gator")
	c := http.Client{}
	res, err := c.Do(req)
	if err != nil {
		return feed, err
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return feed, err
	}

	err = xml.Unmarshal(b, feed)
	if err != nil {
		return feed, err
	}

	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
	}

	return feed, nil
}

func HandlerAgg(s *Status, cmd Command, user database.User) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("no interval given")
	}
	timeBetweenReqs, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return err
	}
	fmt.Printf("Collecting feeds every %v\n", cmd.Args[0])
	ticker := time.NewTicker(timeBetweenReqs)
	for ; ; <-ticker.C {
		err = scrapeFeeds(s, user)
		if err != nil {
			return err
		}
	}
}

func HandlerAddFeed(s *Status, cmd Command, user database.User) error {
	if len(cmd.Args) < 2 {
		return fmt.Errorf("not enough arguments given")
	}
	name := cmd.Args[0]
	url := cmd.Args[1]

	feed := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	}

	err := s.DB.CreateFeed(context.Background(), feed)
	if err != nil {
		return err
	}
	follow := Command{
		Name: "follow",
		Args: cmd.Args[1:],
	}
	err = HandlerFollow(s, follow, user)
	if err != nil {
		return err
	}

	return nil
}

func HandlerAllFeeds(s *Status, cmd Command) error {
	feeds, err := s.DB.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		user, err := s.DB.GetUserName(context.Background(), feed.UserID)
		if err != nil {
			return err
		}
		fmt.Printf("name: %v\n", feed.Name)
		fmt.Printf("url: %v\n", feed.Url)
		fmt.Printf("user: %v\n", user.Name)
	}

	return nil
}

func HandlerFollow(s *Status, cmd Command, user database.User) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("no URL given")
	}
	feed, err := s.DB.GetFeedsURL(context.Background(), cmd.Args[0])
	if err != nil {
		return err
	}

	feedToFollow := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}
	followedFeed, err := s.DB.CreateFeedFollow(context.Background(), feedToFollow)

	fmt.Printf("username: %v\n", followedFeed.UserName)
	fmt.Printf("feedname: %v\n", followedFeed.FeedName)

	return nil
}

func HandlerFollowing(s *Status, cmd Command, user database.User) error {
	feeds, err := s.DB.GetFeedFollowForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}

	fmt.Printf("user %v is following these feeds:\n", user.Name)
	for _, feed := range feeds {
		fmt.Printf("\t%v\n", feed.FeedName)
	}
	return nil
}

func HandlerUnfollow(s *Status, cmd Command, user database.User) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("no feed url given")
	}
	feed, err := s.DB.GetFeedsURL(context.Background(), cmd.Args[0])
	if err != nil {
		return err
	}
	unfollow := database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}
	return s.DB.DeleteFeedFollow(context.Background(), unfollow)
}

func scrapeFeeds(s *Status, user database.User) error {
	feed, err := s.DB.GetNextFeedToFetch(context.Background(), user.ID)
	if err != nil {
		return err
	}
	feedToFetch := database.MarkFeedFetchedParams{
		UpdatedAt: time.Now(),
		ID:        feed.ID,
	}
	err = s.DB.MarkFeedFetched(context.Background(), feedToFetch)
	if err != nil {
		return err
	}

	feedFetched, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		return err
	}

	for _, item := range feedFetched.Channel.Item {

		description := sql.NullString{
			String: item.Description,
			Valid:  true,
		}
		timePub, _ := time.Parse("RFC1123", item.PubDate)

		publishedAt := sql.NullTime{
			Time:  timePub,
			Valid: true,
		}
		timeNow := time.Now()
		savePost := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   timeNow,
			UpdatedAt:   timeNow,
			Title:       item.Title,
			Url:         item.Link,
			Description: description,
			PublishedAt: publishedAt,
			FeedID:      feed.ID,
		}
		err = s.DB.CreatePost(context.Background(), savePost)
		if err != nil {
			pqErr := new(pq.Error)
			if errors.As(err, &pqErr) {
				if pqErr.Code == "23505" {
					continue
				} else {
					return err
				}
			}
		}
	}

	return nil
}

func HandlerBrowse(s *Status, cmd Command, user database.User) error {
	var limit int32
	if len(cmd.Args) == 0 {
		limit = 2
	} else {
		limitArgs, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return err
		}
		limit = int32(limitArgs)
	}
	getPosts := database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  limit,
	}

	posts, err := s.DB.GetPostsForUser(context.Background(), getPosts)
	if err != nil {
		return err
	}

	for _, post := range posts {
		fmt.Printf("%v\n", post.Title)
		fmt.Printf("%v\n", post.Description)
		fmt.Printf("%v\n", post.Url)
		fmt.Printf("\n")
	}

	return nil
}
