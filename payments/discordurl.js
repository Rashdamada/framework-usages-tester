const axios = require('axios');

const BOT_TOKEN = 'YOUR_BOT_TOKEN_HERE'; // Replace with your bot token

// Discord API endpoint to get current bot user info
const url = 'https://discord.com/api/v10/users/@me';

axios.get(url, {
  headers: {
    'Authorization': `Bot ${BOT_TOKEN}`
  }
})
.then(response => {
  console.log('Bot User Info:', response.data);
})
.catch(error => {
  console.error('Error:', error.response ? error.response.data : error.message);
});
