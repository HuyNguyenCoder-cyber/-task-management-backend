ALTER TABLE tasks
    DROP FOREIGN KEY fk_tasks_project;

ALTER TABLE tasks
    MODIFY project_id BIGINT NULL,
    ADD COLUMN created_by BIGINT NULL AFTER project_id,
    ADD COLUMN due_date DATETIME NULL AFTER assignee_id;

UPDATE tasks
INNER JOIN projects ON projects.id = tasks.project_id
SET tasks.created_by = projects.owner_id
WHERE tasks.created_by IS NULL;

ALTER TABLE tasks
    MODIFY created_by BIGINT NOT NULL;

ALTER TABLE tasks
    ADD CONSTRAINT fk_tasks_project
        FOREIGN KEY (project_id)
        REFERENCES projects(id)
        ON DELETE CASCADE,
    ADD CONSTRAINT fk_tasks_created_by
        FOREIGN KEY (created_by)
        REFERENCES users(id)
        ON DELETE CASCADE;

CREATE INDEX idx_tasks_created_by ON tasks(created_by);
CREATE INDEX idx_tasks_due_date ON tasks(due_date);
