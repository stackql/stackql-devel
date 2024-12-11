from flask import Flask, jsonify, render_template, request
import os

app = Flask(__name__)

# Configure the templates directory
app.template_folder = os.path.join(os.path.dirname(__file__), "templates")

@app.route('/api/v1/users', methods=['GET'])
def get_users():
    """Handles paginated user responses."""
    after = int(request.args.get('after', 0))
    end = 5
    next_page = after + end
    max_page = 45
    
    users = []
    for i in range(after + 1, after + end + 1):
        users.append({
            "id": str(i),
            "status": "ACTIVE",
            "created": "2022-07-17T11:36:23.000Z",
            "activated": "2022-07-17T11:36:23.000Z",
            "statusChanged": "2022-07-17T11:37:06.000Z",
            "lastLogin": "2022-07-17T11:38:36.000Z",
            "lastUpdated": "2022-07-17T11:37:06.000Z",
            "passwordChanged": "2022-07-17T11:37:06.000Z",
            "type": {"id": "oty1l8curaosSnZLn697"},
            "profile": {
                "firstName": "sherlock",
                "lastName": "holmes",
                "mobilePhone": None,
                "secondEmail": None,
                "login": f"sherlockholmes{i}@gmail.com",
                "email": f"sherlockholmes{i}@gmail.com"
            },
            "credentials": {
                "password": {},
                "provider": {"type": "OKTA", "name": "OKTA"}
            },
            "_links": {
                "self": {
                    "href": f"https://dummyorg.okta.com/api/v1/users/{i}"
                }
            }
        })

    next_link = None
    if next_page <= max_page:
        next_link = f'<https://{request.host}/api/v1/users?after={next_page}>; rel="next"'

    response_headers = {
        'Datetime': 'now',
        'Link': next_link or f'<https://{request.host}/api/v1/users?after={after}>; rel="self"'
    }

    return render_template('users_template.json', users=users, headers=response_headers)

@app.route('/api/v1/apps', methods=['GET'])
def get_apps():
    """Returns app data."""
    apps = [
        {
            "id": "0oa3cy8k41YM15j4H5d7",
            "name": "saasure",
            "label": "Okta Admin Console",
            "status": "ACTIVE",
            "lastUpdated": "2021-12-17T15:22:44.000Z",
            "created": "2021-12-17T15:22:41.000Z",
            "visibility": {
                "autoSubmitToolbar": False,
                "hide": {"iOS": False, "web": False},
                "appLinks": {"admin": True}
            },
            "signOnMode": "OPENID_CONNECT",
            "settings": {
                "notifications": {"vpn": {"network": {"connection": "DISABLED"}}}
            },
            "_links": {
                "users": {"href": "https://dev-79923018.okta.com/api/v1/apps/0oa3cy8k41YM15j4H5d7/users"}
            }
        }
    ]
    return render_template('apps_template.json', apps=apps)

if __name__ == '__main__':
    app.run(debug=True)
