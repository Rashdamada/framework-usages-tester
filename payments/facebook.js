const graph = require('fbgraph');

// Set your access token (get it from Facebook Graph API Explorer or your app)
graph.setAccessToken('YOUR_ACCESS_TOKEN_HERE');

// Example: Fetch current user's name and profile picture
const params = {
  fields: 'name,picture'
};

graph.get('me', params, function(err, res) {
  if (err) {
    console.error('Error fetching data:', err);
    return;
  }
  console.log('User name:', res.name);
  console.log('Profile picture URL:', res.picture.data.url);
});
