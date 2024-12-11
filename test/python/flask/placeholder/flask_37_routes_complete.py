
import logging
from flask import Flask, render_template, request

app = Flask(__name__)

# Configure logging
logging.basicConfig(level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s")
logger = logging.getLogger(__name__)

@app.before_request
def log_request_info():
    logger.info(f"Request: {request.method} {request.path} - Query: {request.args}")


@app.route('/v1/example/route_1', methods=['GET'])
def route_1():
    return render_template('route_1_template.json')

@app.route('/v1/example/route_2', methods=['GET'])
def route_2():
    return render_template('route_2_template.json')

@app.route('/v1/example/route_3', methods=['GET'])
def route_3():
    return render_template('route_3_template.json')

@app.route('/v1/example/route_4', methods=['GET'])
def route_4():
    return render_template('route_4_template.json')

@app.route('/v1/example/route_5', methods=['GET'])
def route_5():
    return render_template('route_5_template.json')

@app.route('/v1/example/route_6', methods=['GET'])
def route_6():
    return render_template('route_6_template.json')

@app.route('/v1/example/route_7', methods=['GET'])
def route_7():
    return render_template('route_7_template.json')

@app.route('/v1/example/route_8', methods=['GET'])
def route_8():
    return render_template('route_8_template.json')

@app.route('/v1/example/route_9', methods=['GET'])
def route_9():
    return render_template('route_9_template.json')

@app.route('/v1/example/route_10', methods=['GET'])
def route_10():
    return render_template('route_10_template.json')

@app.route('/v1/example/route_11', methods=['GET'])
def route_11():
    return render_template('route_11_template.json')

@app.route('/v1/example/route_12', methods=['GET'])
def route_12():
    return render_template('route_12_template.json')

@app.route('/v1/example/route_13', methods=['GET'])
def route_13():
    return render_template('route_13_template.json')

@app.route('/v1/example/route_14', methods=['GET'])
def route_14():
    return render_template('route_14_template.json')

@app.route('/v1/example/route_15', methods=['GET'])
def route_15():
    return render_template('route_15_template.json')

@app.route('/v1/example/route_16', methods=['GET'])
def route_16():
    return render_template('route_16_template.json')

@app.route('/v1/example/route_17', methods=['GET'])
def route_17():
    return render_template('route_17_template.json')

@app.route('/v1/example/route_18', methods=['GET'])
def route_18():
    return render_template('route_18_template.json')

@app.route('/v1/example/route_19', methods=['GET'])
def route_19():
    return render_template('route_19_template.json')

@app.route('/v1/example/route_20', methods=['GET'])
def route_20():
    return render_template('route_20_template.json')

@app.route('/v1/example/route_21', methods=['GET'])
def route_21():
    return render_template('route_21_template.json')

@app.route('/v1/example/route_22', methods=['GET'])
def route_22():
    return render_template('route_22_template.json')

@app.route('/v1/example/route_23', methods=['GET'])
def route_23():
    return render_template('route_23_template.json')

@app.route('/v1/example/route_24', methods=['GET'])
def route_24():
    return render_template('route_24_template.json')

@app.route('/v1/example/route_25', methods=['GET'])
def route_25():
    return render_template('route_25_template.json')

@app.route('/v1/example/route_26', methods=['GET'])
def route_26():
    return render_template('route_26_template.json')

@app.route('/v1/example/route_27', methods=['GET'])
def route_27():
    return render_template('route_27_template.json')

@app.route('/v1/example/route_28', methods=['GET'])
def route_28():
    return render_template('route_28_template.json')

@app.route('/v1/example/route_29', methods=['GET'])
def route_29():
    return render_template('route_29_template.json')

@app.route('/v1/example/route_30', methods=['GET'])
def route_30():
    return render_template('route_30_template.json')

@app.route('/v1/example/route_31', methods=['GET'])
def route_31():
    return render_template('route_31_template.json')

@app.route('/v1/example/route_32', methods=['GET'])
def route_32():
    return render_template('route_32_template.json')

@app.route('/v1/example/route_33', methods=['GET'])
def route_33():
    return render_template('route_33_template.json')

@app.route('/v1/example/route_34', methods=['GET'])
def route_34():
    return render_template('route_34_template.json')

@app.route('/v1/example/route_35', methods=['GET'])
def route_35():
    return render_template('route_35_template.json')

@app.route('/v1/example/route_36', methods=['GET'])
def route_36():
    return render_template('route_36_template.json')

@app.route('/v1/example/route_37', methods=['GET'])
def route_37():
    return render_template('route_37_template.json')

if __name__ == '__main__':
    app.run(debug=True)
