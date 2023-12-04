<script setup>
import { ref, computed, nextTick, watch, onMounted } from 'vue'

const list = ref([{text: 'Example', complete: false}]);
const inputValue = ref('');
const hideCompleted = ref(false)
const editingIndex = ref(-1)

watch(list, (newList) => {
  localStorage.setItem('myList', JSON.stringify(newList));
}, { deep: true });

const savedList = localStorage.getItem('myList');
if (savedList) {
  list.value = JSON.parse(savedList);
}

function generateUniqueId() {
  return Date.now().toString(36) + Math.random().toString(36).substr(2);
}

function submission() {
  if (inputValue.value.trim()) {
    if (editingIndex.value >= 0) {
      list.value[editingIndex.value].text = inputValue.value;
      editingIndex.value = -1;
    } else {
      list.value.push({ id: generateUniqueId(), text: inputValue.value, complete: false });
    }
    inputValue.value = '';
  }
}

let isDeleting = false;

async function deleteItem(index) {
  if (isDeleting) return;
  isDeleting = true;
  list.value.splice(index, 1);
  await nextTick();
  isDeleting = false;
}
const inputRef = ref(null);

async function editItem(index) {
  inputValue.value = list.value[index].text;
  editingIndex.value = index;
  await nextTick();
  inputRef.value.select();
}

const filteredList = computed(() => {
  return hideCompleted.value ? list.value.filter(item => !item.complete) : list.value;
});

const switchMode = ref([
  { src: "/brightness.png" },
  { src: "/night-mode.png" }
]);
const modeIndex = ref(1);

function toggleMode() {
  modeIndex.value = modeIndex.value === 0 ? 1 : 0;
}

const darkBool = computed(() => modeIndex.value === 0);

watch(darkBool, (newValue) => {
  if (newValue) {
    document.body.classList.add('dark');
    document.body.classList.remove('light');
  } else {
    document.body.classList.remove('dark');
    document.body.classList.add('light');
  }
});

onMounted(() => {
  document.body.classList.add(modeIndex.value === 0 ? 'dark' : 'light');
});

</script>

<template>
  <header>
    <button @click="toggleMode"><img :src="switchMode[modeIndex].src" class="white-image"></button>
  </header>
  <h2>To do</h2>
  <form @submit.prevent="submission">
    <input ref="inputRef" placeholder="e.g. Walk Odie" v-model="inputValue">
    <button id="add">+</button>
  </form>
  <ul>
    <TransitionGroup name="list" tag="ul">
      <li v-for="(item, index) in filteredList" :key="item.id" @click="item.complete = !item.complete"
          :class="{ complete: item.complete }">{{ item.text }}
          <div>
            <button @click.stop="editItem(index)" class="edit"><img src="/edit.png"></button>
            <button @click.stop="deleteItem(index)" class="delete"><img src="/delete.png"></button>
          </div>
      </li>
  </TransitionGroup>
  </ul>
  <footer>
    <button @click="hideCompleted = !hideCompleted" id="hide">Hide completed</button>
  </footer>
</template>

<style>

img {
  width: 20px;
  height: 20px;
  position: relative;
  left: 4px;
  display: flex;
}

body {
  font-family: "Roboto", sans-serif;
  font-size: larger;
  display: flex;
  min-height: 100vh;
  flex-direction: column;
  align-items: center;
  justify-content: space-between;
  padding: 50px;
  margin: 0px;
  position: relative;
  box-sizing: border-box;
}

.light {
  background-image: linear-gradient(to right, rgb(255, 134, 255), rgb(113, 113, 250));
}

.dark {
  background-image: linear-gradient(to right, rgb(115, 39, 115), rgb(33, 33, 103));
}

body::before {
  content: '';
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-image: linear-gradient(to right, rgb(115, 39, 115), rgb(33, 33, 103));
  z-index: -1;
  transition: opacity 0.5s;
}

body.dark::before {
  opacity: 1;
}

body.light::before {
  opacity: 0;
}

h2 {
  display: flex;
  justify-content: center;
  padding: 20px;
  color: white;
}

.complete {
  text-decoration: line-through;
  color: white;
  background-color: rgb(193, 193, 193);
}

form {
  display: flex;
  flex-direction: row;
  justify-content: center;
}

input {
  font-size: larger;
  width: 50vw;
  max-width: 200px;
  background: transparent;
  border: none;
  border-bottom: 1px solid white;
  color: white;
}

::placeholder {
  color: white;
}

:focus {
  outline: none;
}

ul {
  list-style: none;
  display: flex;
  flex-direction: column;
  margin: 0px;
  padding: 15px 0 0 0;
  gap: 15px;
}

li {
  box-shadow: 3px 3px 3px black;
  border: 0.5px solid black;
  background-color: white;
  padding: 7px;
  width: 65vw;
  max-width: 500px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  user-select: none;
  transition: transform 0.3s ease;
}

li:hover {
  transform: scale(1.05);
}

button {
  background-color: transparent;
  color: white;
  border: none;
  border-bottom: 1px solid white;
  width: 30px;
  height: 30px;
  padding: 0px;
  font-size: larger;
  user-select: none;
  transition: transform 0.3s ease;
}

header > button {
  border: none;
}

.white-image {
  filter: invert(100%) brightness(2);
}

header {
  display: flex;
  justify-content: flex-end;
}

#add, #hide {
  transition: font-size 0.3s ease;
}

#add:hover, #hide:hover {
  font-size: 28px;
}

.edit:hover, .delete:hover {
  transform: scale(1.2);
}

li, li:focus, button, button:focus, .edit, .edit:focus, .delete, .delete:focus {
  cursor: pointer;
}

.list-enter-active, .list-leave-active {
  transition: all 0.3s ease;
}

.list-enter-from {
  opacity: 0;
  transform: translateY(30px);
}

.list-leave-to {
  transform: translateX(50px);
  opacity: 0;
}

button.edit, button.delete {
  color: black;
  margin-left: 10px;
  background-color: #f0f0f0;
  color: black;
  border: 1px solid #ddd;
}

#hide {
  width: auto;
}

footer {
  display: flex;
  justify-content: center;
  align-items: center;
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  width: 100%;
  padding: 10px 0;
  margin-bottom: 40px;
}
</style>