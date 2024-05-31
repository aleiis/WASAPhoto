import axios from '../services/axios.js';

export const checkLiveness = async () => {
	try {
		const response = await axios.get('/liveness');
		return response.status === 200;
	} catch (error) {
		return false;
	}
};
