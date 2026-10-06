"""Apply Cully's local-only embedder configuration to a pinned Mem0 server."""

from pathlib import Path
import sys


def replace_once(source: str, before: str, after: str) -> str:
    count = source.count(before)
    if count != 1:
        raise SystemExit(f"pinned Mem0 source changed: expected one match, found {count}")
    return source.replace(before, after)


def main() -> None:
    path = Path(sys.argv[1])
    source = path.read_text()
    source = replace_once(
        source,
        'BUNDLED_LLM_PROVIDERS = ("openai", "anthropic", "gemini")',
        'BUNDLED_LLM_PROVIDERS = ("lmstudio",)',
    )
    source = replace_once(
        source,
        'BUNDLED_EMBEDDER_PROVIDERS = ("openai", "gemini")',
        'BUNDLED_EMBEDDER_PROVIDERS = ("fastembed",)',
    )
    source = replace_once(
        source,
        '            "collection_name": POSTGRES_COLLECTION_NAME,',
        '            "collection_name": POSTGRES_COLLECTION_NAME,\n'
        '            "embedding_model_dims": 384,',
    )
    source = replace_once(
        source,
        '''    "llm": {
        "provider": "openai",
        "config": {"api_key": OPENAI_API_KEY, "temperature": 0.2, "model": DEFAULT_LLM_MODEL},
    },
    "embedder": {"provider": "openai", "config": {"api_key": OPENAI_API_KEY, "model": DEFAULT_EMBEDDER_MODEL}},''',
        '''    "llm": {
        "provider": "lmstudio",
        "config": {"model": "disabled-infer-false", "lmstudio_base_url": "http://127.0.0.1:9/v1"},
    },
    "embedder": {
        "provider": "fastembed",
        "config": {"model": "BAAI/bge-small-en-v1.5", "embedding_dims": 384},
    },''',
    )
    source = replace_once(source, 'OPENAI_API_KEY = os.environ.get("OPENAI_API_KEY")\n', '')
    source = replace_once(
        source,
        'DEFAULT_LLM_MODEL = os.environ.get("MEM0_DEFAULT_LLM_MODEL", "gpt-5-mini")\n'
        'DEFAULT_EMBEDDER_MODEL = os.environ.get("MEM0_DEFAULT_EMBEDDER_MODEL", "text-embedding-3-small")\n',
        '',
    )
    path.write_text(source)


if __name__ == "__main__":
    main()
