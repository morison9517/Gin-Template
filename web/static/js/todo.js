/* ===========================================================================
   todo.js = トップページだけで使うサンプル

   ▼ 相手は internal/handlers/api.go
       GET    /api/todos       一覧
       POST   /api/todos       追加   {"title": "..."}
       PATCH  /api/todos/:id   完了の切り替え
       DELETE /api/todos/:id   削除

   ▼ 項目名(id / title / is_done)は internal/models/todo.go の
     json タグで決まっている。★ここを変えるときは両方直す。

   ▼ api.get / api.post / showMessage / withBusy は main.js の道具。
     整理券(CSRFトークン)は api が自動で付けるので、ここでは意識しない。

   プロダクトが決まったら、このファイルは丸ごと消してよい。
   =========================================================================== */

const form = document.getElementById("todo-form");
const input = document.getElementById("todo-title");
const list = document.getElementById("todo-list");

// このページ(トップ)以外では要素が無いので、何もしない。
if (form && input && list) {
  loadTodos();

  form.addEventListener("submit", async (event) => {
    event.preventDefault();

    const title = input.value.trim();
    if (!title) return;

    await withBusy(form.querySelector("button"), async () => {
      try {
        await api.post("/api/todos", { title });
        input.value = "";
        await loadTodos();
      } catch (error) {
        showMessage(error.message, "error");
      }
    });
  });
}

async function loadTodos() {
  try {
    const data = await api.get("/api/todos");
    render(data.todos ?? []);
  } catch (error) {
    showMessage(error.message, "error");
  }
}

function render(todos) {
  list.textContent = "";

  if (todos.length === 0) {
    const empty = document.createElement("li");
    empty.className = "todo-empty";
    empty.textContent = "まだ何もありません。";
    list.appendChild(empty);
    return;
  }

  for (const todo of todos) {
    list.appendChild(createItem(todo));
  }
}

function createItem(todo) {
  const item = document.createElement("li");
  item.className = todo.is_done ? "todo-item is-done" : "todo-item";

  const checkbox = document.createElement("input");
  checkbox.type = "checkbox";
  checkbox.checked = todo.is_done;
  checkbox.addEventListener("change", () => toggle(todo.id));

  const text = document.createElement("span");
  text.className = "todo-text";
  text.textContent = todo.title;

  const removeButton = document.createElement("button");
  removeButton.className = "btn btn-ghost btn-small";
  removeButton.textContent = "削除";
  removeButton.addEventListener("click", () => remove(todo.id, removeButton));

  item.append(checkbox, text, removeButton);
  return item;
}

async function toggle(id) {
  try {
    await api.patch(`/api/todos/${id}`);
    await loadTodos();
  } catch (error) {
    showMessage(error.message, "error");
  }
}

async function remove(id, button) {
  await withBusy(button, async () => {
    try {
      await api.delete(`/api/todos/${id}`);
      await loadTodos();
    } catch (error) {
      showMessage(error.message, "error");
    }
  });
}
