CREATE TABLE public.pods
(
    pod_id integer NOT NULL,
    date text,
    title text,
    url text,
    explanation text,
    img bytea,
    CONSTRAINT pk_pod_id PRIMARY KEY (pod_id)
);
