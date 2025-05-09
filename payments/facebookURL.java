String access_token = "YOUR_ACCESS_TOKEN";
JSONObject response_me = FB("me?fields=first_name,picture");
String my_name = response_me.getString("first_name");
String photo_url = response_me.getJSONObject("picture").getJSONObject("data").getString("url");

JSONObject FB(String request) {
  String connector = (request.indexOf("?") == -1) ? "?" : "&";
  String[] response = loadStrings("https://graph.facebook.com/" + request + connector + "access_token=" + access_token);
  String json_string = join(response, "");
  return JSONObject.parse(json_string);
}
