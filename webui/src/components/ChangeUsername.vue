<script>

export default {
  data() {
    return {
      newUsername: '',
      disabled: true,
      message: '',
      success: false,
    };
  },
  methods: {
    onInputUpdate() {
      let bytes = new Blob([this.newUsername]).size;
      if (bytes > 16) {
        this.newUsername = this.newUsername.substring(0, this.newUsername.length - 1);
        bytes = new Blob([this.newUsername]).size;
      }
      if (bytes >= 3 && bytes <= 16) {
        this.disabled = false;
      } else {
        this.disabled = true;
      }
    },
    async changeUsername() {
      if (!this.newUsername) {
        this.message = 'Please, enter a new username with 3 to 16 characters.';
        this.success = false;
        return;
      }

      try {
        const response = await this.$axios.patch(`/users/${localStorage.getItem('userid')}`, {
          username: this.newUsername,
        }, {
          headers: {
            Authorization: `Bearer ${localStorage.getItem('userid')}`,
          },
        });

        if (response && response.status === 200) {
          this.message = `Username updated to ${this.newUsername} successfully.`;
          this.success = true;
          localStorage.setItem('username', this.newUsername);

          /* wait one second and then push to profile */
          setTimeout(() => {
            this.$router.push(`/user/${localStorage.getItem('username')}`);
          }, 1000);
        }
      } catch (e) {
        this.message = 'Error while updating the username: ' + (e.response ? e.response.data.message : e.message);
        this.success = false;
      }
    }
  },
};
</script>

<template>
  <div class="change-username">
    <h2>Change Username</h2>
    <form @submit.prevent="changeUsername">
      <input type="text" v-model="newUsername" @input="onInputUpdate" @keyup.enter="changeUsername" placeholder="New username" required />
      <button type="submit" :disabled="disabled">Update</button>
    </form>
    <div v-if="message" :class="{'success': success, 'error': !success}">
      <p>{{ message }}</p>
    </div>
  </div>
</template>

<style scoped>

.change-username {
  font-family: Arial, Helvetica, sans-serif;
}

form {
  display: flex;
  flex-direction: column;
}

input[type="text"] {
  padding: 10px;
  margin-bottom: 20px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 14px;
}

button {
  padding: 10px;
  background-color: #4a4a4a;
  color: white;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  font-size: 14px;
}

button:hover {
  background-color: #525252;
}

button:disabled {
  cursor: not-allowed;
}

.success {
  color: green;
}

.error {
  color: red;
}

</style>
