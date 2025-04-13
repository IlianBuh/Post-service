package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	_ "github.com/IlianBuh/Post-service/internal/storage"
	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}
type tpost struct {
	postId  int
	userId  int
	header  string
	content string
}

func New(
	user string,
	password string,
	host string,
	port int,
	dbname string,
	timeout int,
) (*Storage, error) {
	const op = "postgres.New"
	conn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?connect_timeout=%d&sslmode=disable",
		user, password, host, port, dbname, timeout,
	)
	// conn := fmt.Sprintf(
	// 	"user=%s password=%s host=%s port=%d dbname=%s connect_timeout=%d ssl-mode=",
	// 	user, password, host, port, dbname, timeout,
	// )

	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fail(op, err)
	}
	return &Storage{db: db}, nil
}

func (s *Storage) Save(
	ctx context.Context,
	userId int,
	header string,
	content string,
	themes []string,
) (int, error) {
	const op = "postgres.Save"
	var (
		err    error
		postId int
		wg     *sync.WaitGroup
	)

	if err = ctx.Err(); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	ctx, cncl := context.WithCancel(ctx)
	defer cncl()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fail(op, err)
	}
	defer tx.Rollback()

	wg.Add(2)
	idCh, resCh, errCh := make(chan int, 1), make(chan int, 1), make(chan error, 2)
	defer func() {
		close(idCh)
		close(resCh)
		close(errCh)
	}()
	go s.savePost(ctx, tx, userId, &header, &content, idCh, errCh, wg)
	go s.saveThemes(ctx, tx, themes, idCh, resCh, errCh, wg)
	defer wg.Wait()
	select {
	case postId = <-resCh:
	case err = <-errCh:
		cncl()
		return 0, fail(op, err)
	case <-ctx.Done():
		return 0, fail(op, ctx.Err())
	}

	err = tx.Commit()
	if err != nil {
		return 0, fail(op, err)
	}
	return postId, nil
}

// func (s *Storage) Update(
// 	ctx context.Context,
// 	postId int,
// 	userId int,
// 	header string,
// 	content string,
// 	themes []string,
// ) (int, error) {
// 	const (
// 		op         = "postgres.Update"
// 		selectPost = `
// 			SELECT *
// 			FROM   posts
// 			WHERE post_id=$1`
// 		update = `
// 			UPDATE posts SET header=$1, content=$2`
// 	)

// 	tx, err := s.db.BeginTx(ctx, nil)
// 	if err != nil {
// 		return 0, fail(op, err)
// 	}
// 	defer tx.Rollback()

// 	row := tx.QueryRowContext(ctx, selectPost, postId)
// 	post := tpost{}
// 	if err = row.Scan(&post.postId, &post.userId, &post.header, &post.content); err != nil {

// 		if errors.Is(err, sql.ErrNoRows) {
// 			return 0, fail(op, storage.ErrNotFound)
// 		} else {
// 			return 0, fail(op, err)
// 		}

// 	}
// 	if userId != post.userId {
// 		return 0, fail(op, storage.ErrNotCreator)
// 	}

// 	if header != "" {
// 		post.header = header
// 	}
// 	if content != "" {
// 		post.content = content
// 	}

// 	_, err = tx.ExecContext(ctx, update, post.content, post.header)
// 	if err != nil {
// 		return 0, fail(op, err)
// 	}

// 	// TODO : fix themes

// }

// fail assembles a new error with define structure
// Error message has pattern 'op':'err'
func fail(op string, err error) error {
	return fmt.Errorf("%s: %w", op, err)
}

func (*Storage) loadThemeIds(ctx context.Context, tx *sql.Tx, themes []string) ([]int, error) {
	const op = "postgres.LoadThemeIds"
	const selectThemeQuery = `
		SELECT theme_id
		FROM themes
		WHERE theme_name=$1`
	const insertNewTheme = `
		INSERT INTO themes(theme_name) 
		VALUES ($1)
		RETURNING theme_id`

	slctStmt, err := tx.PrepareContext(ctx, selectThemeQuery)
	if err != nil {
		return nil, fail(op, err)
	}
	defer slctStmt.Close()

	insrtStmt, err := tx.PrepareContext(ctx, insertNewTheme)
	if err != nil {
		return nil, fail(op, err)
	}
	defer insrtStmt.Close()

	themeIds := make([]int, len(themes))
	for i, theme := range themes {
		var id int
		row := slctStmt.QueryRowContext(ctx, theme)

		if err = row.Scan(&id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				if err = insrtStmt.QueryRowContext(ctx, theme).Scan(&id); err != nil {
					return nil, fail(op, err)
				}
			} else {
				return nil, fail(op, err)
			}
		}

		themeIds[i] = id
	}

	return themeIds, nil
}

func (*Storage) savePost(
	ctx context.Context,
	tx *sql.Tx,
	userId int,
	header *string,
	content *string,
	idCh chan int,
	errCh chan error,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	const op = "postgres.savePost"
	const insertNewPost = `
		INSERT INTO posts(user_id, header, content)
		VALUES($1, $2, $3)
		RETURNING post_id`

	insrtStmt, err := tx.PrepareContext(ctx, insertNewPost)
	if err != nil {
		idCh <- 0
		errCh <- fail(op, err)
		return
	}
	defer insrtStmt.Close()

	var postId int
	row := insrtStmt.QueryRowContext(ctx, userId, *header, *content)
	if err = row.Scan(&postId); err != nil {
		idCh <- 0
		errCh <- fail(op, err)
		return
	}
}

func (s *Storage) saveThemes(
	ctx context.Context,
	tx *sql.Tx,
	themes []string,
	idCh <-chan int,
	resCh chan int,
	errCh chan error,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	const (
		op                      = "storage.saveThemes"
		insertPostThemeRelation = `
			INSERT INTO post_theme(post_id, theme_id)
			VALUES($1, $2)`
	)
	var (
		postId int
	)

	themeIds, err := s.loadThemeIds(ctx, tx, themes)
	if err != nil {
		errCh <- fail(op, err)
		return
	}

	insrtStmt, err := tx.PrepareContext(ctx, insertPostThemeRelation)
	if err != nil {
		errCh <- fail(op, err)
		return
	}
	defer insrtStmt.Close()

	select {
	case postId = <-idCh:
	case <-ctx.Done():
		errCh <- ctx.Err()
		return
	}

	for _, themeId := range themeIds {

		_, err = insrtStmt.ExecContext(ctx, postId, themeId)
		if err != nil {
			errCh <- fail(op, err)
			return
		}

	}

	resCh <- postId
}
