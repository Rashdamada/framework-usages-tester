const axios = require('axios');
const accountSid = 'YOUR_TWILIO_ACCOUNT_SID';
const authToken = 'YOUR_TWILIO_AUTH_TOKEN';

const to = '+15558675310';      // Recipient's phone number
const from = '+15017122661';    // Your Twilio phone number
const body = 'Hello from Twilio via direct API call!';

const url = `https://api.twilio.com/2010-04-01/Accounts/${accountSid}/Messages.json`;

axios.post(url, new URLSearchParams({
  To: to,
  From: from,
  Body: body
}), {
  auth: {
    username: accountSid,
    password: authToken
  },
  headers: {
    'Content-Type': 'application/x-www-form-urlencoded'
  }
})
.then(response => {
  console.log('Message SID:', response.data.sid);
})
.catch(error => {
  console.error('Error sending SMS:', error.response ? error.response.data : error.message);
});
