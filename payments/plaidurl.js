const axios = require('axios');

const PLAID_CLIENT_ID = 'YOUR_PLAID_CLIENT_ID';
const PLAID_SECRET = 'YOUR_PLAID_SECRET';
const ACCESS_TOKEN = 'USER_ACCESS_TOKEN';

async function getAccounts() {
  try {
    const response = await axios.post('https://sandbox.plaid.com/accounts/get', {
      client_id: PLAID_CLIENT_ID,
      secret: PLAID_SECRET,
      access_token: ACCESS_TOKEN
    });
    console.log(response.data.accounts);
  } catch (err) {
    console.error(err.response.data);
  }
}

getAccounts();
