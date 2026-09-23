FROM python:3.13-slim AS runtime-base

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1
WORKDIR /app
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt \
    && useradd --uid 65532 --create-home buddy
COPY buddy_service.py ./
USER 65532:65532
EXPOSE 8080

FROM runtime-base AS tests
USER root
COPY tests ./tests
ENV BUDDY_DATABASE_URL=postgresql://buddy:unused@127.0.0.1:5432/buddy
RUN python -m unittest discover -s tests -v

FROM runtime-base AS runtime
CMD ["python", "buddy_service.py"]
