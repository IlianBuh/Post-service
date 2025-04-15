package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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

	postId, err = s.save(ctx, tx, userId, &header, &content, themes)
	if err != nil {
		return 0, fail(op, err)
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
// 		op = "postgres.Update"
// 	)
// 	var (
// 		err error
// 		wg  *sync.WaitGroup
// 	)
// 	if err := ctx.Err(); err != nil {
// 		return 0, fail(op, err)
// 	}
// 	ctx, cncl := context.WithCancel(ctx)
// 	defer cncl()

// 	tx, err := s.db.BeginTx(ctx, nil)
// 	if err != nil {
// 		return 0, fail(op, err)
// 	}
// 	defer tx.Rollback()

// 	wg.Add(2)
// 	idCh, resCh, errCh := make(chan int, 1), make(chan int, 1), make(chan error, 2)
// 	defer func() {
// 		wg.Wait()
// 		close(idCh)
// 		close(resCh)
// 		close(errCh)
// 	}()
// 	go func() {
// 		defer wg.Done()
// 		s.updatePost(ctx, tx, postId, userId, &header, &content, idCh, errCh)
// 	}()
// 	go func() {
// 		defer wg.Done()
// 		s.updateThemes(ctx, tx, themes, idCh, resCh, errCh)
// 	}()
// 	select {}
// 	// TODO : fix themes

// }

// fail assembles a new error with define structure
// Error message has pattern 'op':'err'
func fail(op string, err error) error {
	return fmt.Errorf("%s: %w", op, err)
}

// save saves new post, themes and all new relations
func (s *Storage) save(
	ctx context.Context,
	tx *sql.Tx,
	userId int,
	header *string,
	content *string,
	themes []string,
) (int, error) {
	const (
		op = "storage.saveThemes"
	)
	sendErr := func(err error) (int, error) {
		return 0, fail(op, err)
	}
	ctx, cncl := context.WithCancel(ctx)
	defer cncl()

	postId, thmIds, err := s.fetchAllIds(ctx, tx, userId, header, content, themes)
	if err != nil {
		return sendErr(err)
	}

	err = s.savePostThemeRelations(ctx, tx, postId, thmIds)
	if err != nil {
		return sendErr(err)
	}

	return postId, nil
}

// fetchAllIds fetches ids of the post and all themes. under hood it parallels fetching of  post's id and theme's ids
func (s *Storage) fetchAllIds(
	ctx context.Context,
	tx *sql.Tx,
	userId int,
	header *string,
	content *string,
	themes []string,
) (postId int, thmIds []int, err error) {
	const op = "fetchAllIds"

	sendErr := func(err error) (int, []int, error) {
		return 0, nil, fail(op, err)
	}

	thmIds, err = s.loadThemeIds(ctx, tx, themes)
	if err != nil {
		return sendErr(err)
	}

	postId, err = s.savePost(ctx, tx, userId, header, content)
	if err != nil {
		return sendErr(err)
	}

	return postId, thmIds, nil
}

// savePost saves new post and returns post id
func (s *Storage) savePost(
	ctx context.Context,
	tx *sql.Tx,
	userId int,
	header *string,
	content *string,
) (postId int, err error) {
	const (
		op            = "postgres.savePost"
		insertNewPost = `
			INSERT INTO posts(user_id, header, content)
			VALUES($1, $2, $3)
			RETURNING post_id`
	)
	sendErr := func(err error) (int, error) {
		return 0, fail(op, err)
	}

	insrtStmt, err := tx.PrepareContext(ctx, insertNewPost)
	if err != nil {
		return sendErr(err)
	}
	defer insrtStmt.Close()

	row := insrtStmt.QueryRowContext(ctx, userId, *header, *content)
	if err = row.Scan(&postId); err != nil {
		return sendErr(err)
	}

	return postId, nil
}

//	 loadThemeIds loads id by theme names from slice. If name does not exist
//		in database new theme is created and the id of the new theme is returned
func (*Storage) loadThemeIds(
	ctx context.Context,
	tx *sql.Tx,
	themes []string,
) (themeIds []int, err error) {
	const (
		op               = "postgres.LoadThemeIds"
		selectThemeQuery = `
			SELECT theme_id
			FROM themes
			WHERE theme_name=$1;`
		insertNewTheme = `
			INSERT INTO themes(theme_name) 
			VALUES ($1)
			RETURNING theme_id`
	)

	sendErr := func(err error) ([]int, error) {
		return nil, fail(op, err)
	}

	slctStmt, err := tx.PrepareContext(ctx, selectThemeQuery)
	if err != nil {
		return sendErr(err)
	}
	defer slctStmt.Close()

	insrtStmt, err := tx.PrepareContext(ctx, insertNewTheme)
	if err != nil {
		return sendErr(err)
	}
	defer insrtStmt.Close()

	themeIds = make([]int, len(themes))
	var id int
	for i, theme := range themes {
		select {
		case <-ctx.Done():
			return sendErr(ctx.Err())
		default:
		}

		row := slctStmt.QueryRowContext(ctx, theme)

		if err = row.Scan(&id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				if err = insrtStmt.QueryRowContext(ctx, theme).Scan(&id); err != nil {
					return sendErr(err)
				}
			} else {
				return sendErr(err)
			}
		}

		themeIds[i] = id
	}

	return themeIds, nil
}

func (*Storage) savePostThemeRelations(
	ctx context.Context,
	tx *sql.Tx,
	postId int,
	thmIds []int,
) error {
	const (
		op                      = "postgres.savePostThemeRelations"
		insertPostThemeRelation = `
			INSERT INTO post_theme(post_id, theme_id)
			VALUES($1, $2)`
	)
	sendErr := func(err error) error {
		return fail(op, err)
	}

	insrtStmt, err := tx.PrepareContext(ctx, insertPostThemeRelation)
	if err != nil {
		return sendErr(err)
	}
	defer insrtStmt.Close()

	for _, themeId := range thmIds {

		_, err = insrtStmt.ExecContext(ctx, postId, themeId)
		if err != nil {
			return sendErr(err)
		}

	}

	return nil
}

// func (s *Storage) updatePost(
// 	ctx context.Context,
// 	tx *sql.Tx,
// 	postId int,
// 	userId int,
// 	header *string,
// 	content *string,
// 	idCh chan int,
// 	errCh chan error,
// ) {
// 	const (
// 		op         = "postgres.updatePost"
// 		selectPost = `
// 			SELECT *
// 			FROM   posts
// 			WHERE post_id=$1`
// 		update = `
// 			UPDATE posts SET header=$1, content=$2`
// 	)
// 	var err error

// 	row := tx.QueryRowContext(ctx, selectPost, postId)
// 	post := tpost{}
// 	if err = row.Scan(&post.postId, &post.userId, &post.header, &post.content); err != nil {

// 		if errors.Is(err, sql.ErrNoRows) {
// 			idCh <- 0
// 			s.tryWError(errCh, fail(op, storage.ErrNotFound))
// 		} else {
// 			idCh <- 0
// 			s.tryWError(errCh, fail(op, err))
// 		}

// 	}
// 	if userId != post.userId {
// 		idCh <- 0
// 		s.tryWError(errCh, fail(op, storage.ErrNotCreator))
// 	}

// 	if *header != "" {
// 		post.header = *header
// 	}
// 	if *content != "" {
// 		post.content = *content
// 	}

// 	_, err = tx.ExecContext(ctx, update, post.content, post.header)
// 	if err != nil {
// 		idCh <- 0
// 		s.tryWError(errCh, fail(op, err))
// 	}

// }

// func (s *Storage) updateThemes(
// 	ctx context.Context,
// 	tx *sql.Tx,
// 	themes []string,
// 	idCh chan int,
// 	resCh chan int,
// 	errCh chan error,
// ) {
// 	const (
// 		op           = "postgres.updateThemes"
// 		deleteThemes = `
// 			DELETE FROM post_theme WHERE post_id=$1`
// 	)
// 	var (
// 		postId int
// 		err    error
// 		wg     *sync.WaitGroup
// 	)

// 	wg.Add(1)
// 	_idCh, _resCh := make(chan int, 1), make(chan int, 1)
// 	defer func() {
// 		wg.Wait()
// 		close(_idCh)
// 		close(_resCh)
// 	}()
// 	go func() {
// 		defer wg.Done()
// 		s.saveThemes(ctx, tx, themes, _idCh, _resCh, errCh)
// 	}()

// 	select {
// 	case postId = <-idCh:
// 	case err = <-errCh:
// 		s.tryWError(errCh, fail(op, err))
// 		return
// 	case <-ctx.Done():
// 		s.tryWError(errCh, fail(op, ctx.Err()))
// 		return
// 	}

// 	_, err = tx.ExecContext(ctx, deleteThemes, postId)
// 	if err != nil {
// 		s.tryWError(errCh, fail(op, err))
// 		return
// 	}

// 	select {
// 	case _idCh <- postId:
// 	default:
// 		s.tryWError(errCh, fail(op, storage.ErrBlockedChannel))
// 		return
// 	}

// }

func readChan[T any](
	ch chan T,
	errCh chan error,
) (T, error) {
	var res T

	select {
	case res = <-ch:
	case err := <-errCh:
		return res, err
	}

	return res, nil
}

// func (s *Storage) T() {
// 	tx, err := s.db.Begin()
// 	if err != nil {
// 		panic(err.Error())
// 	}

// 	// stmt, err := tx.Prepare(`SELECT * FROM themes`)
// 	// if err != nil {
// 	// 	panic("stmt" + err.Error())
// 	// }

// 	rows := tx.QueryRow("SELECT * FROM themes")
// 	stms2, err := tx.Prepare(`INSERT INTO post_theme(post_id, theme_id) VALUES($1, $1) `)
// 	if err != nil {
// 		panic("stms2" + err.Error())
// 	}
// 	if err != nil {
// 		panic("rows" + err.Error())
// 	}
// 	_ = rows
// 	// for rows.Next() {
// 	// }

// 	rows2, err := stms2.Query(0)
// 	if err != nil {
// 		panic("rows2" + err.Error())
// 	}
// 	_ = rows2
// 	tx.Commit()
// }
