CREATE TABLE IF NOT EXISTS public.user
(
	id serial NOT NULL PRIMARY key,
	name text NOT NULL
)
TABLESPACE pg_default;

CREATE TABLE IF NOT EXISTS public.product
(
	id serial NOT NULL PRIMARY key,
	name text NOT NULL,
    quantity integer NOT NULL CHECK (quantity >= 0)
)
TABLESPACE pg_default;

CREATE TABLE IF NOT EXISTS public.ordering
(
	id serial NOT NULL PRIMARY key,
    user_id serial NOT NULL,
    product_id serial NOT NULL,
    quantity integer NOT NULL CHECK (quantity >= 0),
    CONSTRAINT fk_ordering_user
      FOREIGN KEY(user_id) 
	  REFERENCES "user"(id),
    CONSTRAINT fk_ordering_product
      FOREIGN KEY(product_id) 
	  REFERENCES product(id)
)
TABLESPACE pg_default;