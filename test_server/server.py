import os

from app import app

if __name__ == "__main__":
    # get application config from environment
    for envvar in ["VULNERABLE", "SECRET"]:
        if envvar in os.environ:
            app.config[envvar] = os.environ[envvar]

    if os.environ.get("USE_GEVENT"):
        from gevent import monkey

        monkey.patch_all()
        from gevent.pywsgi import WSGIServer

        WSGIServer(
            (
                "0.0.0.0",
                5000,
            ),
            app.wsgi_app,
        ).serve_forever()
    else:
        app.run("0.0.0.0", 5000)
