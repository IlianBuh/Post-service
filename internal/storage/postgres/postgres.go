package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/IlianBuh/Post-service/internal/domain/models"
	"github.com/IlianBuh/Post-service/internal/storage"
	"github.com/IlianBuh/Post-service/internal/storage/events"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}
type record struct {
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
	sendErr := func(err error) (int, error) {
		return 0, fail(op, err)
	}

	if err = ctx.Err(); err != nil {
		return sendErr(err)
	}
	ctx, cncl := context.WithCancel(ctx)
	defer cncl()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return sendErr(err)
	}
	defer tx.Rollback()

	postId, err = s.save(ctx, tx, userId, &header, &content, themes)
	if err != nil {
		return sendErr(err)
	}

	payload := events.CollectEventPayload(userId, header)
	eventId := events.CollectEventId(userId)
	err = s.saveEvent(ctx, tx, eventId, events.TypeCteated, payload)
	if err != nil {
		return sendErr(err)
	}

	err = tx.Commit()
	if err != nil {
		return sendErr(err)
	}

	return postId, nil
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

// savePostThemeRelations make notes for matching themes to post
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

// saveEvent saves new event
func (s *Storage) saveEvent(
	ctx context.Context,
	tx *sql.Tx,
	eventId string,
	eventType string,
	payload string,
) error {
	const (
		op        = "postgres.saveEvent"
		insrtStmt = `
		INSERT INTO events(event_id, type, payload)
		VALUES ($1, $2, $3);
		`
	)

	_, err := tx.ExecContext(ctx, insrtStmt, eventId, eventType, payload)
	if err != nil {
		return fail(op, err)
	}

	return nil
}

func (s *Storage) Update(
	ctx context.Context,
	postId int,
	userId int,
	header string,
	content string,
	themes []string,
) (int, error) {
	const op = "postgres.Update"
	var (
		err error
		rec record
	)
	sendErr := func(err error) (int, error) {
		return 0, fail(op, err)
	}

	if err = ctx.Err(); err != nil {
		return sendErr(err)
	}
	ctx, cncl := context.WithCancel(ctx)
	defer cncl()

	rec, err = s.makePostRec(ctx, postId, userId, header, content)
	if err != nil {
		return sendErr(err)
	}

	err = s.update(ctx, rec, themes)
	if err != nil {
		return sendErr(err)
	}

	return postId, nil
}

// update updates all information related to the post that has the postId
func (s *Storage) update(
	ctx context.Context,
	rec record,
	themes []string,
) error {
	const (
		op = "postgres.update"
	)
	sendErr := func(err error) error {
		return fail(op, err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return sendErr(err)
	}
	defer tx.Rollback()

	err = s.updatePost(ctx, tx, &rec)
	if err != nil {
		return sendErr(err)
	}

	err = s.updateThemes(ctx, tx, rec.postId, themes)
	if err != nil {
		return sendErr(err)
	}

	err = tx.Commit()
	if err != nil {
		return sendErr(err)
	}

	return nil
}

// makePostRec creates a record for new post information
func (s *Storage) makePostRec(
	ctx context.Context,
	postId int,
	userId int,
	header string,
	content string,
) (record, error) {
	const (
		op = "postgres.newPostRec"
	)
	sendErr := func(err error) (record, error) {
		return record{}, fail(op, err)
	}

	rec, err := s.post(ctx, postId)
	if err != nil {
		return sendErr(err)
	}

	if !s.isCreator(rec.userId, userId) {
		return sendErr(storage.ErrNotCreator)
	}

	if header != "" {
		rec.header = header
	}
	if content != "" {
		rec.content = content
	}

	return rec, nil
}

// updatePost updates post records with replacing content and header
func (s *Storage) updatePost(
	ctx context.Context,
	tx *sql.Tx,
	post *record,
) error {
	const (
		op        = "postgres.updatePost"
		updtQuery = `
			UPDATE posts SET header=$1, content=$2 WHERE post_id=$3`
	)

	_, err := tx.ExecContext(ctx, updtQuery, post.header, post.content, post.postId)
	if err != nil {
		return fail(op, err)
	}

	return nil
}

// updateThemes updates list of themes that match to post with postId
func (s *Storage) updateThemes(
	ctx context.Context,
	tx *sql.Tx,
	postId int,
	themes []string,
) error {
	const (
		op = "postgres.updateThemes"
	)
	sendErr := func(err error) error {
		return fail(op, err)
	}

	thmIds, err := s.loadThemeIds(ctx, tx, themes)
	if err != nil {
		return sendErr(err)
	}

	err = s.deleteRelations(ctx, tx, postId)
	if err != nil {
		return sendErr(err)
	}

	err = s.savePostThemeRelations(ctx, tx, postId, thmIds)
	if err != nil {
		return sendErr(err)
	}

	return nil
}

func (s *Storage) Delete(
	ctx context.Context,
	postId int,
	userId int,
) error {
	const op = "postgres.Delete"
	sendErr := func(err error) error {
		return fail(op, err)
	}

	ctx, cncl := context.WithCancel(ctx)
	defer cncl()

	rec, err := s.post(ctx, postId)
	if err != nil {
		return sendErr(err)
	}

	if !s.isCreator(rec.userId, userId) {
		return sendErr(storage.ErrNotCreator)
	}

	err = s.delete(ctx, postId)
	if err != nil {
		return sendErr(err)
	}

	return nil
}

// delete deletes all information related to post with the postId
func (s *Storage) delete(
	ctx context.Context,
	postId int,
) error {
	const (
		op = "postgres.delete"
	)
	sendErr := func(err error) error {
		return fail(op, err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return sendErr(err)
	}
	defer tx.Rollback()

	err = s.deletePost(ctx, tx, postId)
	if err != nil {
		return sendErr(err)
	}

	err = s.deleteRelations(ctx, tx, postId)
	if err != nil {
		return sendErr(err)
	}

	err = tx.Commit()
	if err != nil {
		return sendErr(err)
	}

	return nil
}

// deletePost deletes post with the postId
func (s *Storage) deletePost(
	ctx context.Context,
	tx *sql.Tx,
	postId int,
) error {
	const (
		op       = "postgres.deletePost"
		dltQuery = `
			DELETE FROM posts WHERE post_id=$1;
		`
	)

	_, err := tx.ExecContext(ctx, dltQuery, postId)
	if err != nil {
		return fail(op, err)
	}

	return nil
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

// loadThemeIds loads id by theme names from slice. If name does not exist
// in database new theme is created and the id of the new theme is returned
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

// post returns record with posts' information
func (s *Storage) post(
	ctx context.Context,
	postId int,
) (record, error) {
	const (
		op        = "postgres.post"
		slctQuery = `
			SELECT post_id, user_id, header, content
			FROM posts
			WHERE post_id = $1;
		`
	)
	var rec record

	row := s.db.QueryRowContext(ctx, slctQuery, postId)
	if err := row.Scan(&rec.postId, &rec.userId, &rec.header, &rec.content); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return rec, fail(op, storage.ErrNotFound)
		}

		return rec, fail(op, err)
	}

	return rec, nil
}

func (s *Storage) deleteRelations(
	ctx context.Context,
	tx *sql.Tx,
	postId int,
) error {
	const (
		op               = "postgres.deleteRelations"
		dltRelationQuery = `
			DELETE FROM post_theme WHERE post_id=$1;
		`
	)

	_, err := tx.ExecContext(ctx, dltRelationQuery, postId)
	if err != nil {
		return fail(op, err)
	}

	return nil
}

// isCreator checks is the user creator of the record
func (s *Storage) isCreator(recUserId, userId int) bool {
	return recUserId == userId
}

// Stop stops working of storage entity
func (s *Storage) Stop() error {
	const (
		op = "postgres.Stop"
	)
	err := s.db.Close()
	if err != nil {
		return fail(op, storage.ErrClose)
	}

	return nil
}

// EventPage returns page of events. Page is a slice of limit size
func (s *Storage) EventPage(ctx context.Context, limit int) ([]models.Event, error) {
	const (
		op        = "postgres.EventPage"
		slctQuery = `
		SELECT id, event_id, type, payload
		FROM events
		WHERE status != 'done' AND reserved_to < (NOW() AT TIME ZONE 'UTC-3')
		LIMIT $1`
	)
	var (
		events  []models.Event = make([]models.Event, 0, limit)
		sendErr                = func(err error) ([]models.Event, error) { return nil, fail(op, err) }
	)

	rows, err := s.db.QueryContext(ctx, slctQuery, limit)
	if err != nil {
		return sendErr(err)
	}
	defer rows.Close()

	var event models.Event
	for rows.Next() {
		if err = rows.Scan(&event.Id, &event.EventId, &event.Type, &event.Payload); err != nil {
			return sendErr(err)
		}

		events = append(events, event)
	}

	if len(events) == 0 {
		return sendErr(storage.ErrNoEvents)
	}

	return events, nil
}

// Reserve reserves events with id from ids list
func (s *Storage) Reserve(ctx context.Context, ids []int) error {
	const (
		op        = "postgres.Reserve"
		rsrvQuery = `
		Update events 
		SET reserved_to=((NOW() AT TIME ZONE 'UTC-3') + INTERVAL '5 minutes' )
		WHERE id = ANY ($1)
		`
	)

	_, err := s.db.ExecContext(ctx, rsrvQuery, pq.Array(ids))
	if err != nil {
		return fail(op, err)
	}

	return nil
}

// DeleteEvent deletes events with id from ids list
func (s *Storage) DeleteEvent(ctx context.Context, ids []int) error {
	const (
		op       = "postgres.Delete"
		dltQuery = `
		UPDATE events SET status='done' WHERE id = ANY($1)
		`
	)

	_, err := s.db.ExecContext(ctx, dltQuery, pq.Array(ids))
	if err != nil {
		return fail(op, err)
	}

	return nil
}

// fail assembles a new error with define structure
// Error message has pattern 'op':'err'
func fail(op string, err error) error {
	return fmt.Errorf("%s: %w", op, err)
}
