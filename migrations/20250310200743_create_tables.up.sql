CREATE TABLE users (
                       user_id BIGINT NOT NULL,
                       PRIMARY KEY (user_id)
);


CREATE TABLE links(
                      link_id BIGSERIAL,
                      link_url TEXT NOT NULL UNIQUE,
                      last_update_check TIMESTAMP,
                      PRIMARY KEY (link_id)
);


CREATE TABLE tags (
                      tag_id BIGSERIAL,
                      tag_name VARCHAR(50),

                      PRIMARY KEY (tag_id)
);

CREATE TABLE userLinks(
                          user_id BIGINT NOT null,
                          link_id BIGINT NOT NULL,
                          tag_id  BIGINT DEFAULT NULL,

                          PRIMARY KEY(user_id, link_id),
                          FOREIGN KEY (user_id)
                              REFERENCES users(user_id),

                          FOREIGN KEY (link_id)
                              REFERENCES links(link_id),

                          FOREIGN KEY (tag_id)
                            REFERENCES tags(tag_id)
);
