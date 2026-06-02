DROP INDEX idx_tasks_due_date ON tasks;
DROP INDEX idx_tasks_created_by ON tasks;

ALTER TABLE tasks
    DROP FOREIGN KEY fk_tasks_created_by,
    DROP FOREIGN KEY fk_tasks_project;

ALTER TABLE tasks
    DROP COLUMN due_date,
    DROP COLUMN created_by;

DELETE FROM tasks WHERE project_id IS NULL;

ALTER TABLE tasks
    MODIFY project_id BIGINT NOT NULL,
    ADD CONSTRAINT fk_tasks_project
        FOREIGN KEY (project_id)
        REFERENCES projects(id)
        ON DELETE CASCADE;
