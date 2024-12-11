from flask import Flask, request, render_template, jsonify
import os

app = Flask(__name__)
app.template_folder = "templates"

# Routes generated from mockserver configuration

@app.route('/', methods=['POST'])
def route_1():
    return render_template('template_1.json')

@app.route('/', methods=['POST'])
def route_2():
    return render_template('template_2.json')

@app.route('/', methods=['POST'])
def route_3():
    return render_template('template_3.json')

@app.route('/', methods=['POST'])
def route_4():
    return render_template('template_4.json')

@app.route('/', methods=['POST'])
def route_5():
    return render_template('template_5.json')

@app.route('/', methods=['POST'])
def route_6():
    return render_template('template_6.json')

@app.route('/2013-04-01/hostedzone/A00000001AAAAAAAAAAAA/rrset', methods=['GET'])
def route_7():
    return render_template('template_7.json')

@app.route('/2013-04-01/hostedzone/some-id/rrset/', methods=['POST'])
def route_8():
    return render_template('template_8.json')

@app.route('/2013-04-01/hostedzone/some-id/rrset/', methods=['POST'])
def route_9():
    return render_template('template_9.json')

@app.route('/', methods=['POST'])
def route_10():
    return render_template('template_10.json')

@app.route('/', methods=['GET'])
def route_11():
    return render_template('template_11.json')

@app.route('/', methods=['GET'])
def route_12():
    return render_template('template_12.json')

@app.route('/', methods=['GET'])
def route_13():
    return render_template('template_13.json')

@app.route('/', methods=['GET'])
def route_14():
    return render_template('template_14.json')

@app.route('/', methods=['POST'])
def route_15():
    return render_template('template_15.json')

@app.route('/', methods=['POST'])
def route_16():
    return render_template('template_16.json')

@app.route('/', methods=['POST'])
def route_17():
    return render_template('template_17.json')

@app.route('/', methods=['POST'])
def route_18():
    return render_template('template_18.json')

@app.route('/', methods=['POST'])
def route_19():
    return render_template('template_19.json')

@app.route('/', methods=['POST'])
def route_20():
    return render_template('template_20.json')

@app.route('/', methods=['POST'])
def route_21():
    return render_template('template_21.json')

@app.route('/', methods=['POST'])
def route_22():
    return render_template('template_22.json')

@app.route('/', methods=['POST'])
def route_23():
    return render_template('template_23.json')

@app.route('/', methods=['POST'])
def route_24():
    return render_template('template_24.json')

@app.route('/', methods=['POST'])
def route_25():
    return render_template('template_25.json')

@app.route('/', methods=['POST'])
def route_26():
    return render_template('template_26.json')

@app.route('/', methods=['POST'])
def route_27():
    return render_template('template_27.json')

@app.route('/', methods=['POST'])
def route_28():
    return render_template('template_28.json')

@app.route('/', methods=['POST'])
def route_29():
    return render_template('template_29.json')

@app.route('/', methods=['POST'])
def route_30():
    return render_template('template_30.json')

@app.route('/', methods=['POST'])
def route_31():
    return render_template('template_31.json')

@app.route('/', methods=['POST'])
def route_32():
    return render_template('template_32.json')

@app.route('/', methods=['POST'])
def route_33():
    return render_template('template_33.json')

@app.route('/', methods=['POST'])
def route_34():
    return render_template('template_34.json')

@app.route('/', methods=['POST'])
def route_35():
    return render_template('template_35.json')

@app.route('/', methods=['POST'])
def route_36():
    return render_template('template_36.json')

@app.route('/', methods=['POST'])
def route_37():
    return render_template('template_37.json')

@app.route('/', methods=['POST'])
def route_38():
    return render_template('template_38.json')

@app.route('/', methods=['POST'])
def route_39():
    return render_template('template_39.json')

@app.route('/', methods=['POST'])
def route_40():
    return render_template('template_40.json')

@app.route('/', methods=['POST'])
def route_41():
    return render_template('template_41.json')

@app.route('/', methods=['POST'])
def route_42():
    return render_template('template_42.json')

@app.route('/', methods=['POST'])
def route_43():
    return render_template('template_43.json')

@app.route('/', methods=['POST'])
def route_44():
    return render_template('template_44.json')

@app.route('/', methods=['POST'])
def route_45():
    return render_template('template_45.json')

@app.route('/', methods=['POST'])
def route_46():
    return render_template('template_46.json')

@app.route('/', methods=['POST'])
def route_47():
    return render_template('template_47.json')

@app.route('/', methods=['POST'])
def route_48():
    return render_template('template_48.json')

@app.route('/', methods=['POST'])
def route_49():
    return render_template('template_49.json')

@app.route('/', methods=['POST'])
def route_50():
    return render_template('template_50.json')

@app.route('/', methods=['POST'])
def route_51():
    return render_template('template_51.json')

@app.route('/', methods=['POST'])
def route_52():
    return render_template('template_52.json')

@app.route('/', methods=['POST'])
def route_53():
    return render_template('template_53.json')

@app.route('/', methods=['POST'])
def route_54():
    return render_template('template_54.json')

@app.route('/', methods=['POST'])
def route_55():
    return render_template('template_55.json')

@app.route('/', methods=['POST'])
def route_56():
    return render_template('template_56.json')

@app.route('/', methods=['POST'])
def route_57():
    return render_template('template_57.json')

@app.route('/', methods=['POST'])
def route_58():
    return render_template('template_58.json')

@app.route('/', methods=['POST'])
def route_59():
    return render_template('template_59.json')

@app.route('/', methods=['POST'])
def route_60():
    return render_template('template_60.json')

@app.route('/', methods=['POST'])
def route_61():
    return render_template('template_61.json')

@app.route('/', methods=['POST'])
def route_62():
    return render_template('template_62.json')

@app.route('/', methods=['POST'])
def route_63():
    return render_template('template_63.json')

@app.route('/', methods=['POST'])
def route_64():
    return render_template('template_64.json')

@app.route('/', methods=['POST'])
def route_65():
    return render_template('template_65.json')

@app.route('/', methods=['POST'])
def route_66():
    return render_template('template_66.json')

@app.route('/', methods=['POST'])
def route_67():
    return render_template('template_67.json')

@app.route('/', methods=['POST'])
def route_68():
    return render_template('template_68.json')

@app.route('/', methods=['POST'])
def route_69():
    return render_template('template_69.json')

@app.route('/', methods=['POST'])
def route_70():
    return render_template('template_70.json')

@app.route('/', methods=['POST'])
def route_71():
    return render_template('template_71.json')

@app.route('/', methods=['POST'])
def route_72():
    return render_template('template_72.json')

@app.route('/', methods=['POST'])
def route_73():
    return render_template('template_73.json')

@app.route('/', methods=['POST'])
def route_74():
    return render_template('template_74.json')

@app.route('/', methods=['POST'])
def route_75():
    return render_template('template_75.json')

@app.route('/', methods=['POST'])
def route_76():
    return render_template('template_76.json')

@app.route('/', methods=['POST'])
def route_77():
    return render_template('template_77.json')

@app.route('/', methods=['POST'])
def route_78():
    return render_template('template_78.json')

@app.route('/', methods=['POST'])
def route_79():
    return render_template('template_79.json')

@app.route('/', methods=['POST'])
def route_80():
    return render_template('template_80.json')

@app.route('/', methods=['POST'])
def route_81():
    return render_template('template_81.json')

if __name__ == "__main__":
    app.run(debug=True, host="0.0.0.0", port=5000)
