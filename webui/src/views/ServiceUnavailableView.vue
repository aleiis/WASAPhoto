<script>
import {checkLiveness} from "../utils/livenessCheck.js";

export default {
	mounted() {
		checkLiveness()
		.then(() => {
			this.$router.push('/')
		})
	}
}

</script>

<template>
	<div class="notfound">
		<div class="header">
			<img class="logo" src="/src/assets/logo.png"/>
			<h1>WASAPhoto</h1>
		</div>
		<h3>We're sorry, but the service is currently unavailable. Please try again later.</h3>
		<button @click="this.$router.push('/')">Retry</button>
	</div>
</template>

<style scoped>

.notfound {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 50px;
	padding-top: 5%;
	background-color: #f5f5f5;
	height: calc(100% - 121px);
	font-family: Arial, Helvetica, sans-serif
}

.header {
	display: flex;
	flex-direction: row;
	justify-content: center;
	align-items: center;
	margin-bottom: 50px;
}

img {
	width: 100px;
	height: 100px;
	margin-right: 40px;
}

h1 {
	font-family: "PoetsenOne";
}

button {
	background-color: transparent;
	border: none;
	color: rgb(120, 120, 120);
	padding: 15px 32px;
	text-align: center;
	text-decoration: none;
	display: inline-block;
	font-size: 16px;
	margin: 4px 2px;
	cursor: pointer;
}

button:hover {
	color: rgb(90, 90, 90);
	text-decoration: underline;
}

</style>
