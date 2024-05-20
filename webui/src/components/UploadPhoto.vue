<script>

export default {
  data() {
    return {
      selectedFile: null,
      disabled: true,
      message: '',
      success: false,
    };
  },
  methods: {
    onFileChange(event) {
      this.selectedFile = event.target.files[0];
      this.disabled = this.selectedFile.type.includes('image') ? false : true;
    },
    async uploadPhoto() {
      if (!this.selectedFile) {
        this.message = 'Please, select a photo to upload.';
        this.success = false;
        return;
      }

      if (!await this.selectedFile.type.includes('image')) {
        this.message = 'Please, select an image file.';
        this.success = false;
        return;
      }

      try {
        const response = await this.$axios.post(`/users/${localStorage.getItem('userid')}/photos/`, this.selectedFile, {
          headers: {
            'Content-Type': this.selectedFile.type,
            'Authorization': `Bearer ${localStorage.getItem('userid')}`
          },
        });

        if (response && response.status === 201) {
          this.message = 'Photo uploaded successfully.';
          this.success = true;

          /* wait two seconds and then push to profile */
          setTimeout(() => {
            this.$router.push(`/user/${localStorage.getItem('username')}`);
          }, 2000);
        }
      } catch (e) {
        this.message = 'Error uploading photo: ' + (e.response ? e.response.data.message : e.message);
        this.success = false;
      }
    }
  },
};
</script>

<template>
  <div class="upload-photo">
    <h2>Upload a photo</h2>
    <form @submit.prevent="uploadPhoto">
      <input type="file" @change="onFileChange" accept="image/png, image/jpeg" required />
      <button type="submit" :disabled="disabled">Upload</button>
    </form>
    <div v-if="message" :class="{'success': success, 'error': !success}">
      <p>{{ message }}</p>
    </div>
  </div>
</template>

<style scoped>

.upload-photo {
  font-family: Arial, Helvetica, sans-serif;
}

form {
  display: flex;
  flex-direction: column;
}

input[type="file"] {
  font-size: 14px;
  margin-bottom: 15px;
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
