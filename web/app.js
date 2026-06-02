const API_BASE = "";
const TOKEN_KEY = "access_token";
const USER_KEY = "auth_user";
let taskRealtimeSocket = null;

function token() {
  return localStorage.getItem(TOKEN_KEY);
}

function currentUser() {
  try {
    return JSON.parse(localStorage.getItem(USER_KEY) || "null");
  } catch {
    return null;
  }
}

function currentUserId() {
  const user = currentUser();
  const tokenPayload = parseJwt(token() || "");
  return Number(user?.id || tokenPayload?.user_id || 0);
}

function setStatus(message, type = "") {
  const el = document.querySelector("#status");
  if (!el) return;
  el.textContent = message || "";
  el.className = `status ${type}`.trim();
  el.hidden = !message;
}

async function api(path, options = {}) {
  const headers = {
    "Content-Type": "application/json",
    ...(options.headers || {}),
  };

  if (token()) {
    headers.Authorization = `Bearer ${token()}`;
  }

  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  });

  const payload = await res.json().catch(() => ({}));
  if (res.status === 401) {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
  }
  if (!res.ok || payload.success === false) {
    throw new Error(payload.error || payload.message || `Request failed: ${res.status}`);
  }

  return payload.data ?? payload;
}

function formData(form) {
  return Object.fromEntries(new FormData(form).entries());
}

function numberOrNull(value) {
  if (value === undefined || value === null || value === "") return null;
  const n = Number(value);
  return Number.isFinite(n) ? n : null;
}

function isoOrNull(value) {
  if (!value) return null;
  return new Date(value).toISOString();
}

function localDateTimeValue(value) {
  if (!value) return "";
  const date = new Date(value);
  const offsetMs = date.getTimezoneOffset() * 60000;
  return new Date(date.getTime() - offsetMs).toISOString().slice(0, 16);
}

function createTaskPayload(data) {
  return {
    project_id: numberOrNull(data.project_id),
    title: data.title,
    description: data.description || "",
    status: data.status || "todo",
    assignee_id: numberOrNull(data.assignee_id),
    due_date: isoOrNull(data.due_date),
  };
}

function updateTaskPayload(data) {
  return {
    title: data.title,
    description: data.description || "",
    status: data.status || "todo",
    due_date: isoOrNull(data.due_date),
  };
}

function projectPayload(data) {
  return {
    name: data.name,
    description: data.description || "",
  };
}

function escapeHtml(value) {
  return String(value ?? "").replace(/[&<>"']/g, (char) => {
    return {
      "&": "&amp;",
      "<": "&lt;",
      ">": "&gt;",
      '"': "&quot;",
      "'": "&#039;",
    }[char];
  });
}

function queryParam(name) {
  return new URLSearchParams(window.location.search).get(name);
}

function fmtDate(value) {
  if (!value) return "-";
  return new Date(value).toLocaleString();
}

function requireLogin() {
  if (!token()) {
    window.location.href = "login.html";
    return false;
  }
  return true;
}

function setupNav() {
  const loggedIn = Boolean(token());
  document.querySelectorAll("[data-auth='in']").forEach((el) => {
    el.hidden = !loggedIn;
  });
  document.querySelectorAll("[data-auth='out']").forEach((el) => {
    el.hidden = loggedIn;
  });

  const logout = document.querySelector("#logout");
  if (logout) {
    logout.addEventListener("click", () => {
      localStorage.removeItem(TOKEN_KEY);
      localStorage.removeItem(USER_KEY);
      window.location.href = "login.html";
    });
  }
}

function wsURL(path) {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${protocol}//${window.location.host}${path}`;
}

function connectTaskRealtime(onRealtimeEvent) {
  if (typeof onRealtimeEvent !== "function") return () => {};

  let reconnectTimer = null;
  let closedByClient = false;

  const connect = () => {
    if (closedByClient) return;
    const socket = new WebSocket(wsURL("/ws"));
    taskRealtimeSocket = socket;

    socket.onmessage = async (event) => {
      try {
        const raw = typeof event.data === "string"
          ? event.data
          : (event.data && typeof event.data.text === "function" ? await event.data.text() : "{}");
        const payload = JSON.parse(raw || "{}");
        if (payload.event) onRealtimeEvent(payload);
      } catch {
        // Ignore invalid realtime payload.
      }
    };

    socket.onclose = () => {
      if (closedByClient) return;
      reconnectTimer = setTimeout(connect, 1500);
    };
  };

  connect();

  return () => {
    closedByClient = true;
    if (reconnectTimer) clearTimeout(reconnectTimer);
    if (taskRealtimeSocket && taskRealtimeSocket.readyState <= 1) {
      taskRealtimeSocket.close();
    }
    taskRealtimeSocket = null;
  };
}

function updateTaskStatusInList(taskId, status, containerId = "task-list") {
  const statusBadge = document.querySelector(
    `#${containerId} [data-task-id="${taskId}"] [data-task-status-badge="true"]`,
  );
  if (!statusBadge) return false;
  statusBadge.textContent = status;
  return true;
}

function updateTaskStatusInDetail(taskId, status) {
  const badge = document.querySelector(`#task-detail [data-task-id="${taskId}"] [data-task-status-badge="true"]`);
  if (!badge) return false;
  badge.textContent = status;
  return true;
}

function renderTaskCard(task, userId, canUseTaskForm) {
  const isCreator = Number(task.created_by) === userId;
  const isAssignee = Number(task.assignee_id || 0) === userId;
  const projectLabel = task.project_id ? `<a href="project-detail.html?id=${task.project_id}">${task.project_id}</a>` : "No project";
  return `
    <article class="item" data-task-id="${task.id}">
      <h3>${escapeHtml(task.title)}</h3>
      <p class="muted">${escapeHtml(task.description || "No description")}</p>
      <div class="kv"><span>ID</span><strong>${task.id}</strong></div>
      <div class="kv"><span>Project</span><span>${projectLabel}</span></div>
      <div class="kv"><span>Status</span><span class="badge" data-task-status-badge="true">${escapeHtml(task.status)}</span></div>
      <div class="kv"><span>Creator</span><span>${task.created_by ?? "-"}</span></div>
      <div class="kv"><span>Assignee</span><span>${task.assignee_id || "-"}</span></div>
      <div class="kv"><span>Due</span><span>${fmtDate(task.due_date)}</span></div>
      <div class="actions">
        <a class="button secondary" href="task-detail.html?id=${task.id}">View</a>
        ${isCreator && canUseTaskForm ? `<button class="secondary" data-edit-task="${task.id}">Edit</button>` : ""}
        ${isAssignee && !isCreator ? `<button class="secondary" data-status-task="${task.id}">Update status</button>` : ""}
        ${isCreator ? `<button class="danger" data-delete-task="${task.id}">Delete</button>` : ""}
      </div>
    </article>
  `;
}

async function loadTasks(containerId = "task-list", filterProjectId = null) {
  const list = document.querySelector(`#${containerId}`);
  if (!list) return [];
  list.innerHTML = "<div class=\"empty\">Loading tasks...</div>";

  const userId = currentUserId();
  const tasks = await api(filterProjectId ? `/tasks?project_id=${encodeURIComponent(filterProjectId)}` : "/tasks");
  const filtered = tasks;

  if (!filtered.length) {
    list.innerHTML = "<div class=\"empty\">No tasks found.</div>";
    return filtered;
  }

  const canUseTaskForm = Boolean(document.querySelector("#task-form [name='id']"));
  list.innerHTML = filtered.map((task) => renderTaskCard(task, userId, canUseTaskForm)).join("");

  list.querySelectorAll("[data-delete-task]").forEach((button) => {
    button.addEventListener("click", async () => {
      if (!confirm("Delete this task?")) return;
      await api(`/tasks/${button.dataset.deleteTask}`, { method: "DELETE" });
      setStatus("Task deleted.", "ok");
      await loadTasks(containerId, filterProjectId);
    });
  });

  list.querySelectorAll("[data-edit-task]").forEach((button) => {
    button.addEventListener("click", async () => {
      const task = filtered.find((item) => String(item.id) === button.dataset.editTask);
      fillTaskForm(task);
      setStatus(`Editing task #${task.id}.`, "ok");
    });
  });

  list.querySelectorAll("[data-status-task]").forEach((button) => {
    button.addEventListener("click", async () => {
      const task = filtered.find((item) => String(item.id) === button.dataset.statusTask);
      const status = prompt("New status: todo, in_progress, done", task?.status || "todo");
      if (!status) return;
      await api(`/tasks/${button.dataset.statusTask}`, {
        method: "PUT",
        body: JSON.stringify({ status }),
      });
      setStatus("Task status updated.", "ok");
      await loadTasks(containerId, filterProjectId);
    });
  });

  return filtered;
}

function upsertTaskCardFromEvent(eventPayload, containerId = "task-list", prepend = true) {
  const list = document.querySelector(`#${containerId}`);
  if (!list) return false;

  const taskId = Number(eventPayload?.task_id || 0);
  if (taskId <= 0) return false;

  const existing = list.querySelector(`[data-task-id="${taskId}"]`);
  if (existing) return false;

  const emptyEl = list.querySelector(".empty");
  if (emptyEl) {
    list.innerHTML = "";
  }

  const task = {
    id: taskId,
    project_id: eventPayload.project_id ?? null,
    title: eventPayload.title || `Task #${taskId}`,
    description: eventPayload.description || "",
    status: eventPayload.status || "todo",
    created_by: Number(eventPayload.created_by || 0) || null,
    assignee_id: eventPayload.assignee_id ?? null,
    due_date: eventPayload.due_date || null,
  };

  const html = renderTaskCard(task, currentUserId(), Boolean(document.querySelector("#task-form [name='id']")));
  list.insertAdjacentHTML(prepend ? "afterbegin" : "beforeend", html);
  return true;
}

function fillTaskForm(task) {
  const form = document.querySelector("#task-form");
  if (!form || !task) return;
  form.elements.id.value = task.id || "";
  form.elements.project_id.value = task.project_id || "";
  form.elements.title.value = task.title || "";
  form.elements.description.value = task.description || "";
  form.elements.status.value = task.status || "todo";
  form.elements.due_date.value = localDateTimeValue(task.due_date);
}

async function populateProjectSelect(selectId = "task-project-select") {
  const select = document.querySelector(`#${selectId}`);
  if (!select) return [];
  const projects = await api("/projects");
  select.innerHTML = `<option value="">No project</option>` + projects.map((project) => (
    `<option value="${project.id}">${escapeHtml(project.name)}</option>`
  )).join("");
  return projects;
}

async function initTasksPage() {
  if (!requireLogin()) return;
  await populateProjectSelect();
  const user = currentUser();
  const assigneeDisplay = document.querySelector("[name='assignee_display']");
  if (assigneeDisplay) {
    assigneeDisplay.value = user?.full_name || user?.email || `User #${currentUserId()}`;
  }
  await loadTasks();
  const disconnectRealtime = connectTaskRealtime(async (event) => {
    if (event.event === "task_updated") {
      const taskId = Number(event?.task_id || 0);
      const status = String(event?.status || "");
      if (!(taskId > 0 && status)) return;
      const ok = updateTaskStatusInList(taskId, status, "task-list");
      console.debug("[WS][tasks] update", { taskId, status, updated: ok });
    } else if (event.event === "task_created") {
      const ok = upsertTaskCardFromEvent(event, "task-list", true);
      console.debug("[WS][tasks] create", { taskId: event.task_id, inserted: ok });
    }
  });
  window.addEventListener("beforeunload", disconnectRealtime, { once: true });

  const form = document.querySelector("#task-form");
  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    const data = formData(form);
    const id = data.id;
    const saved = id
      ? await api(`/tasks/${id}`, { method: "PUT", body: JSON.stringify(updateTaskPayload(data)) })
      : await api("/tasks", { method: "POST", body: JSON.stringify(createTaskPayload(data)) });
    form.reset();
    setStatus(`Task ${id ? "updated" : "created"}: #${saved.id}`, "ok");
    await loadTasks();
  });

  document.querySelector("#clear-task-form").addEventListener("click", () => {
    form.reset();
    form.elements.id.value = "";
    if (assigneeDisplay) {
      assigneeDisplay.value = user?.full_name || user?.email || `User #${currentUserId()}`;
    }
    setStatus("");
  });
}

async function initTaskDetailPage() {
  if (!requireLogin()) return;
  const id = queryParam("id");
  if (!id) {
    setStatus("Missing task id in URL.", "error");
    return;
  }

  const renderTaskDetail = (task) => {
    document.querySelector("#task-detail").innerHTML = `
      <div data-task-id="${task.id}">
      <h2>${escapeHtml(task.title)}</h2>
      <p>${escapeHtml(task.description || "No description")}</p>
      <div class="kv"><span>ID</span><strong>${task.id}</strong></div>
      <div class="kv"><span>Project</span><span>${task.project_id ? `<a href="project-detail.html?id=${task.project_id}">${task.project_id}</a>` : "No project"}</span></div>
      <div class="kv"><span>Status</span><span class="badge" data-task-status-badge="true">${escapeHtml(task.status)}</span></div>
      <div class="kv"><span>Creator</span><span>${task.created_by}</span></div>
      <div class="kv"><span>Assignee</span><span>${task.assignee_id || "-"}</span></div>
      <div class="kv"><span>Due</span><span>${fmtDate(task.due_date)}</span></div>
      <div class="kv"><span>Created</span><span>${fmtDate(task.created_at)}</span></div>
      <div class="kv"><span>Updated</span><span>${fmtDate(task.updated_at)}</span></div>
      </div>
    `;
  };

  const renderCommentItem = (comment) => {
    const username = comment.username || comment.user_name || `User #${comment.user_id || "?"}`;
    return `
      <article class="comment-item">
        <div class="comment-author">${escapeHtml(username)}:</div>
        <div>${escapeHtml(comment.content || "")}</div>
      </article>
    `;
  };

  const loadComments = async () => {
    const commentList = document.querySelector("#comment-list");
    if (!commentList) return;
    commentList.innerHTML = "<div class=\"empty\">Đang tải bình luận...</div>";
    const comments = await api(`/tasks/${id}/comments`);
    if (!comments.length) {
      commentList.innerHTML = "<div class=\"empty\">Chưa có bình luận nào.</div>";
      return;
    }

    commentList.innerHTML = comments.map((comment) => renderCommentItem(comment)).join("");
  };

  const appendComment = (comment) => {
    const commentList = document.querySelector("#comment-list");
    if (!commentList) return;
    const isEmpty = commentList.querySelector(".empty");
    if (isEmpty) {
      commentList.innerHTML = "";
    }
    commentList.insertAdjacentHTML("beforeend", renderCommentItem(comment));
  };

  renderTaskDetail(await api(`/tasks/${id}`));
  await loadComments();

  const sendCommentButton = document.querySelector("#btn-send-comment");
  const commentInput = document.querySelector("#comment-input");
  if (sendCommentButton && commentInput) {
    sendCommentButton.addEventListener("click", async () => {
      const content = String(commentInput.value || "").trim();
      if (!content) {
        setStatus("Vui lòng nhập nội dung comment.", "error");
        return;
      }

      await api(`/tasks/${id}/comments`, {
        method: "POST",
        body: JSON.stringify({ content }),
      });
      commentInput.value = "";
      setStatus("Đã gửi comment.", "ok");
    });
  }

  const disconnectRealtime = connectTaskRealtime(async (event) => {
    if (event.event === "task_updated") {
      if (String(event.task_id) !== String(id)) return;
      if (event.status) {
        const ok = updateTaskStatusInDetail(event.task_id, event.status);
        console.debug("[WS][task-detail] update", { taskId: event.task_id, status: event.status, updated: ok });
      }
      const fresh = await api(`/tasks/${id}`);
      renderTaskDetail(fresh);
    } else if (event.event === "comment_created") {
      if (String(event.task_id) !== String(id)) return;
      appendComment({
        username: event.username,
        content: event.content,
        user_id: event.user_id,
      });
    }
  });
  window.addEventListener("beforeunload", disconnectRealtime, { once: true });
}

async function loadProjects() {
  const list = document.querySelector("#project-list");
  if (!list) return [];
  list.innerHTML = "<div class=\"empty\">Loading projects...</div>";

  const projects = await api("/projects");
  if (!projects.length) {
    list.innerHTML = "<div class=\"empty\">No projects found.</div>";
    return projects;
  }

  const userId = currentUserId();
  list.innerHTML = projects.map((project) => {
    const isOwner = Number(project.owner_id) === userId;
    return `
    <article class="item">
      <h3>${escapeHtml(project.name)}</h3>
      <p class="muted">${escapeHtml(project.description || "No description")}</p>
      <div class="kv"><span>ID</span><strong>${project.id}</strong></div>
      <div class="kv"><span>Owner</span><span>${project.owner_id}</span></div>
      <div class="actions">
        <a class="button secondary" href="project-detail.html?id=${project.id}">Open</a>
        ${isOwner ? `<button class="secondary" data-edit-project="${project.id}">Edit</button>` : ""}
        ${isOwner ? `<button class="danger" data-delete-project="${project.id}">Delete</button>` : ""}
      </div>
    </article>
  `}).join("");

  list.querySelectorAll("[data-delete-project]").forEach((button) => {
    button.addEventListener("click", async () => {
      if (!confirm("Delete this project?")) return;
      await api(`/projects/${button.dataset.deleteProject}`, { method: "DELETE" });
      setStatus("Project deleted.", "ok");
      await loadProjects();
    });
  });

  list.querySelectorAll("[data-edit-project]").forEach((button) => {
    button.addEventListener("click", async () => {
      const project = projects.find((item) => String(item.id) === button.dataset.editProject);
      fillProjectForm(project);
      setStatus(`Editing project #${project.id}.`, "ok");
    });
  });

  return projects;
}

function fillProjectForm(project) {
  const form = document.querySelector("#project-form");
  if (!form || !project) return;
  form.elements.id.value = project.id || "";
  form.elements.name.value = project.name || "";
  form.elements.description.value = project.description || "";
}

async function initProjectsPage() {
  if (!requireLogin()) return;
  await loadProjects();

  const form = document.querySelector("#project-form");
  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    const data = formData(form);
    const id = data.id;
    const saved = id
      ? await api(`/projects/${id}`, { method: "PUT", body: JSON.stringify(projectPayload(data)) })
      : await api("/projects", { method: "POST", body: JSON.stringify(projectPayload(data)) });
    form.reset();
    setStatus(`Project ${id ? "updated" : "created"}: #${saved.id}`, "ok");
    await loadProjects();
  });

  document.querySelector("#clear-project-form").addEventListener("click", () => {
    form.reset();
    form.elements.id.value = "";
    setStatus("");
  });
}

async function initProjectDetailPage() {
  if (!requireLogin()) return;
  const id = queryParam("id");
  if (!id) {
    setStatus("Missing project id in URL.", "error");
    return;
  }

  const userId = currentUserId();
  const project = await api(`/projects/${id}`);
  const isOwner = Number(project.owner_id) === userId;
  document.querySelector("#project-detail").innerHTML = `
    <h2>${escapeHtml(project.name)}</h2>
    <p>${escapeHtml(project.description || "No description")}</p>
    <div class="kv"><span>ID</span><strong>${project.id}</strong></div>
    <div class="kv"><span>Owner</span><span>${project.owner_id}</span></div>
    <div class="kv"><span>Created</span><span>${fmtDate(project.created_at)}</span></div>
    <div class="kv"><span>Updated</span><span>${fmtDate(project.updated_at)}</span></div>
  `;

  let members = await loadProjectMembers(id, isOwner);
  populateAssigneeSelect(members);

  const memberForm = document.querySelector("#member-form");
  memberForm.hidden = !isOwner;
  memberForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const data = formData(memberForm);
    await api(`/projects/${id}/members`, {
      method: "POST",
      body: JSON.stringify({ email: data.email }),
    });
    memberForm.reset();
    setStatus("Member added.", "ok");
    members = await loadProjectMembers(id, isOwner);
    populateAssigneeSelect(members);
  });

  const taskForm = document.querySelector("#task-form");
  taskForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const saved = await api(`/projects/${id}/tasks`, {
      method: "POST",
      body: JSON.stringify(createTaskPayload(formData(taskForm))),
    });
    taskForm.reset();
    populateAssigneeSelect(members);
    setStatus(`Task created in this project: #${saved.id}`, "ok");
    await loadTasks("project-task-list", id);
  });

  await loadTasks("project-task-list", id);
  const disconnectRealtime = connectTaskRealtime(async (event) => {
    if (event.event === "task_updated") {
      const taskId = Number(event?.task_id || 0);
      const status = String(event?.status || "");
      if (!(taskId > 0 && status)) return;
      const ok = updateTaskStatusInList(taskId, status, "project-task-list");
      console.debug("[WS][project-detail] update", { taskId, status, updated: ok });
    } else if (event.event === "task_created") {
      if (String(event.project_id || "") !== String(id)) return;
      const ok = upsertTaskCardFromEvent(event, "project-task-list", true);
      console.debug("[WS][project-detail] create", { taskId: event.task_id, inserted: ok });
    }
  });
  window.addEventListener("beforeunload", disconnectRealtime, { once: true });
}

async function loadProjectMembers(projectId, isOwner) {
  const membersEl = document.querySelector("#project-members");
  membersEl.innerHTML = "<div class=\"empty\">Loading members...</div>";
  const members = await api(`/projects/${projectId}/members`);
  if (!members.length) {
    membersEl.innerHTML = "<div class=\"empty\">No members found.</div>";
    return members;
  }

  membersEl.innerHTML = members.map((member) => `
    <article class="item compact">
      <h3>${escapeHtml(member.full_name || member.email)}</h3>
      <p class="muted">${escapeHtml(member.email)}</p>
      <div class="kv"><span>ID</span><strong>${member.id}</strong></div>
      <div class="kv"><span>Role</span><span class="badge">${escapeHtml(member.role)}</span></div>
    </article>
  `).join("");

  document.querySelector("#member-form").hidden = !isOwner;
  return members;
}

function populateAssigneeSelect(members) {
  const select = document.querySelector("#project-task-assignee");
  if (!select) return;
  select.innerHTML = `<option value="">Unassigned / me</option>` + members.map((member) => (
    `<option value="${member.id}">${escapeHtml(member.full_name || member.email)} (${escapeHtml(member.role)})</option>`
  )).join("");
}

async function initRegisterPage() {
  const form = document.querySelector("#register-form");
  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    const data = formData(form);
    const user = await api("/auth/register", {
      method: "POST",
      body: JSON.stringify({
        email: data.email,
        password: data.password,
        full_name: data.full_name,
      }),
    });
    setStatus(`Registered ${user.email}. You can log in now.`, "ok");
    form.reset();
  });
}

async function initLoginPage() {
  const form = document.querySelector("#login-form");
  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    const data = formData(form);
    const auth = await api("/auth/login", {
      method: "POST",
      body: JSON.stringify({
        email: data.email,
        password: data.password,
      }),
    });
    localStorage.setItem(TOKEN_KEY, auth.access_token);
    localStorage.setItem(USER_KEY, JSON.stringify(auth.user));
    setStatus("Logged in. Redirecting to tasks...", "ok");
    setTimeout(() => {
      window.location.href = "tasks.html";
    }, 500);
  });
}

async function initProfilePage() {
  if (!requireLogin()) return;
  const user = currentUser();
  const tokenPayload = parseJwt(token());
  const userId = user?.id || tokenPayload?.user_id;

  document.querySelector("#profile-summary").innerHTML = `
    <h2>${escapeHtml(user?.full_name || "Logged-in user")}</h2>
    <div class="kv"><span>User ID</span><strong>${userId || "-"}</strong></div>
    <div class="kv"><span>Email</span><span>${escapeHtml(user?.email || "-")}</span></div>
    <div class="kv"><span>Token expiry</span><span>${tokenPayload?.exp ? fmtDate(tokenPayload.exp * 1000) : "-"}</span></div>
  `;

  if (userId) {
    try {
      const fresh = await api(`/users/${userId}`);
      document.querySelector("#profile-api").innerHTML = `
        <h3>Profile from /users/${userId}</h3>
        <div class="kv"><span>Full name</span><span>${escapeHtml(fresh.full_name)}</span></div>
        <div class="kv"><span>Email</span><span>${escapeHtml(fresh.email)}</span></div>
        <div class="kv"><span>Created</span><span>${fmtDate(fresh.created_at)}</span></div>
      `;
    } catch (error) {
      document.querySelector("#profile-api").innerHTML = `
        <div class="empty">Could not load /users/${userId}: ${escapeHtml(error.message)}</div>
      `;
    }
  }
}

function parseJwt(jwt) {
  try {
    return JSON.parse(atob(jwt.split(".")[1].replace(/-/g, "+").replace(/_/g, "/")));
  } catch {
    return null;
  }
}

function initHomePage() {
  const loggedIn = Boolean(token());
  document.querySelector("#home-state").textContent = loggedIn
    ? "You are logged in. Open Tasks or Projects to exercise the API."
    : "Register or log in to start testing protected endpoints.";
}

window.addEventListener("DOMContentLoaded", async () => {
  setupNav();
  const page = document.body.dataset.page;
  try {
    if (page === "home") initHomePage();
    if (page === "register") await initRegisterPage();
    if (page === "login") await initLoginPage();
    if (page === "tasks") await initTasksPage();
    if (page === "task-detail") await initTaskDetailPage();
    if (page === "projects") await initProjectsPage();
    if (page === "project-detail") await initProjectDetailPage();
    if (page === "profile") await initProfilePage();
  } catch (error) {
    setStatus(error.message, "error");
  }
});
