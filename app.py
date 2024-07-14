from flask import Flask, render_template, request
import requests

app = Flask(__name__)

# URL of the REST API server
REST_API_URL = 'http://127.0.0.1:8888/'

@app.route('/', methods=['GET', 'POST'])
def index():
    if request.method == 'POST':
        prompt = request.form['prompt']

        if not prompt:
            return render_template('index.html', error='Prompt is required')

        # Prepare the request data
        data = {'prompt': prompt}

        try:
            # Send POST request to the REST API server
            response = requests.post(REST_API_URL, json=data)
            response.raise_for_status()  # Raise exception for bad status codes

            # Parse the response JSON
            result = response.json()

            return render_template('index.html', prompt=prompt, completion=result['completion'], execution_time=result['execution_time'])

        except requests.exceptions.RequestException as e:
            error_msg = f'Error sending request to server: {str(e)}'
            return render_template('index.html', error=error_msg)

    return render_template('index.html')

if __name__ == '__main__':
    app.run(debug=True, port=8510)


'''
    conda create -n web python=3.12 -y 
    conda activate web
    
    pip install flask
    python app.py 
    
    pip install py2wasm                       # after 3.11

    py2wasm app.py -o app.wasm
    wasmer run app.wasm
'''