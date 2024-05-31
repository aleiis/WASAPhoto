<script>
import LoginForm from '/src/components/LoginForm.vue'

export default {
	data() {
		return {
			errormsg: ''
		}
	},
	components: {
		LoginForm
	},
	methods: {
		async doLogin(username) {
			try {
				const response = await this.$axios.post("/session", {
					username: username
				});
				localStorage.setItem("userid", response.data);
				localStorage.setItem("username", username);
				this.$router.push({path: '/home'})
			} catch (e) {
				if (e.response && e.response.status === 400) {
					this.errormsg = "Invalid username. The username must consist of 3 to 16 alphanumeric characters.";
				} else if (e.response && e.response.status === 500) {
					this.errormsg = "An internal error occurred. Please try again later.";
				} else {
					this.errormsg = e.toString();
				}
			}
		}
	},
	mounted() {
		const username = localStorage.getItem('username');
		const userid = localStorage.getItem('userid');

		if (username && userid) {
			this.$router.push('/home');
		}
	}
}
</script>

<template>
	<div class="loginview">
		<div class="header">
			<img class="logo" src="/src/assets/logo.png"/>
			<h1>WASAPhoto</h1>
		</div>
		<LoginForm id="login" @login="doLogin"/>
		<p v-if="errormsg" id="error">{{ errormsg }}</p>
	</div>
</template>

<style scoped>

.loginview {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 50px;
	padding-top: 5%;
	background-color: #f5f5f5;
	height: calc(100% - 121px);
}

.loginview .header {
	display: flex;
	flex-direction: row;
	justify-content: center;
	align-items: center;
	margin-bottom: 50px;
}

.loginview .header img {
	width: 100px;
	height: 100px;
	margin-right: 40px;
}

.loginview .header h1 {
	font-family: "PoetsenOne";
}

#error {
	font-family: Arial, Helvetica, sans-serif;
	color: rgb(231, 0, 0);
	font-size: 14px;
	margin-top: 20px;
}

</style>
