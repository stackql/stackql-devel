from flask import Flask, request, render_template, make_response
import os

app = Flask(__name__)
app.template_folder = "templates"

# Routes generated from mockserver configuration

@app.route('/', methods=['POST'])
def route_1():
    response_body = render_template('template_1.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_2():
    response_body = render_template('template_2.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_3():
    response_body = render_template('template_3.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 400, headers)

@app.route('/', methods=['POST'])
def route_4():
    response_body = render_template('template_4.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_5():
    response_body = render_template('template_5.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 400, headers)

@app.route('/', methods=['POST'])
def route_6():
    response_body = render_template('template_6.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/2013-04-01/hostedzone/A00000001AAAAAAAAAAAA/rrset', methods=['GET'])
def route_7():
    response_body = render_template('template_7.json')
    headers = {'content-type': 'text/xml'}
    return make_response(response_body, 200, headers)

@app.route('/2013-04-01/hostedzone/some-id/rrset/', methods=['POST'])
def route_8():
    response_body = render_template('template_8.json')
    headers = {'content-type': 'text/xml'}
    return make_response(response_body, 200, headers)

@app.route('/2013-04-01/hostedzone/some-id/rrset/', methods=['POST'])
def route_9():
    response_body = render_template('template_9.json')
    headers = {'content-type': 'text/xml'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_10():
    response_body = render_template('template_10.json')
    headers = {'content-type': 'text/xml'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['GET'])
def route_11():
    response_body = render_template('template_11.json')
    headers = {'content-type': 'text/xml'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['GET'])
def route_12():
    response_body = render_template('template_12.json')
    headers = {'content-type': 'text/xml'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['GET'])
def route_13():
    response_body = render_template('template_13.json')
    headers = {'content-type': 'text/xml'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['GET'])
def route_14():
    response_body = render_template('template_14.json')
    headers = {'content-type': 'text/xml'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_15():
    response_body = render_template('template_15.json')
    headers = {'content-type': 'text/xml'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_16():
    response_body = render_template('template_16.json')
    headers = {}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_17():
    response_body = render_template('template_17.json')
    headers = {'content-type': 'text/xml'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_18():
    response_body = render_template('template_18.json')
    headers = {'content-type': 'text/xml'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_19():
    response_body = render_template('template_19.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_20():
    response_body = render_template('template_20.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_21():
    response_body = render_template('template_21.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_22():
    response_body = render_template('template_22.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_23():
    response_body = render_template('template_23.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_24():
    response_body = render_template('template_24.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_25():
    response_body = render_template('template_25.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_26():
    response_body = render_template('template_26.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_27():
    response_body = render_template('template_27.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_28():
    response_body = render_template('template_28.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_29():
    response_body = render_template('template_29.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_30():
    response_body = render_template('template_30.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_31():
    response_body = render_template('template_31.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_32():
    response_body = render_template('template_32.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_33():
    response_body = render_template('template_33.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_34():
    response_body = render_template('template_34.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_35():
    response_body = render_template('template_35.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_36():
    response_body = render_template('template_36.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_37():
    response_body = render_template('template_37.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_38():
    response_body = render_template('template_38.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_39():
    response_body = render_template('template_39.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_40():
    response_body = render_template('template_40.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_41():
    response_body = render_template('template_41.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_42():
    response_body = render_template('template_42.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_43():
    response_body = render_template('template_43.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_44():
    response_body = render_template('template_44.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_45():
    response_body = render_template('template_45.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_46():
    response_body = render_template('template_46.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_47():
    response_body = render_template('template_47.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_48():
    response_body = render_template('template_48.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_49():
    response_body = render_template('template_49.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_50():
    response_body = render_template('template_50.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_51():
    response_body = render_template('template_51.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_52():
    response_body = render_template('template_52.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_53():
    response_body = render_template('template_53.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_54():
    response_body = render_template('template_54.json')
    headers = {}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_55():
    response_body = render_template('template_55.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_56():
    response_body = render_template('template_56.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 501, headers)

@app.route('/', methods=['POST'])
def route_57():
    response_body = render_template('template_57.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 403, headers)

@app.route('/', methods=['POST'])
def route_58():
    response_body = render_template('template_58.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_59():
    response_body = render_template('template_59.json')
    headers = {'content-type': 'application/json'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_60():
    response_body = render_template('template_60.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_61():
    response_body = render_template('template_61.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_62():
    response_body = render_template('template_62.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_63():
    response_body = render_template('template_63.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_64():
    response_body = render_template('template_64.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_65():
    response_body = render_template('template_65.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_66():
    response_body = render_template('template_66.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_67():
    response_body = render_template('template_67.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_68():
    response_body = render_template('template_68.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_69():
    response_body = render_template('template_69.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_70():
    response_body = render_template('template_70.json')
    headers = {}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_71():
    response_body = render_template('template_71.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_72():
    response_body = render_template('template_72.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_73():
    response_body = render_template('template_73.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_74():
    response_body = render_template('template_74.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_75():
    response_body = render_template('template_75.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_76():
    response_body = render_template('template_76.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_77():
    response_body = render_template('template_77.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_78():
    response_body = render_template('template_78.json')
    headers = {'content-type': 'application/x-amz-json-1.0'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_79():
    response_body = render_template('template_79.json')
    headers = {'content-type': 'text/xml'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_80():
    response_body = render_template('template_80.json')
    headers = {'content-type': 'text/xml'}
    return make_response(response_body, 200, headers)

@app.route('/', methods=['POST'])
def route_81():
    response_body = render_template('template_81.json')
    headers = {'content-type': 'text/xml'}
    return make_response(response_body, 200, headers)

if __name__ == "__main__":
    app.run(debug=True, host="0.0.0.0", port=5000)
