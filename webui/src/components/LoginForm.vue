<script>
export default {
  data() {
    return {
      username: ''
    }
  },
  methods: {
    login() {
      this.$emit('login', this.username);
    }
  },
  emits: ['login']
}
</script>

<template>
  <div class="loginform">
    <p>Log in to WASAPhoto</p>
    <input v-model="username" @keyup.enter="login" type="text" placeholder="Username">
    <button @click="login">Log In</button>
  </div>
</template>

<style scoped>

.loginform {
  background-color: #fff;
  border-radius: 10px; /* rounded corners */
  box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
  padding: 40px;
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 300px;
  height: auto;
}

p {
  font-family: Arial, sans-serif;
  font-size: 20px;
}

input[type="text"] {
  border: 1px solid #ccc; /* light gray */
  border-radius: 4px;
  font-size: 16px;
  padding: 10px;
  width: 100%;
}

button {
  background-color: #007bff; /* blue */
  border-radius: 8px;
  border: none;
  color: white;
  cursor: pointer;
  font-size: 16px;
  margin-top: 20px;
  margin-bottom: 20px;
  padding: 10px;
  width: 100%;
}

</style>