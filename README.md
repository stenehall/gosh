A minimal dashoard written in Go.

![Dashboard preview](https://raw.githubusercontent.com/stenehall/gosh/gh-pages/assets/screenshot.png)

Here's a tiny example of what a config file might look like.

For each set you can define a name and an icon. 
For each site you can define a name, an url and an icon. If you define an icon gosh will not attempt to download the favicon but instead use the icon provided.

```yaml
title: Gosh dashboard
showtitle: false
port: 8080
sets:
  - name: News
    icon: fa-newspaper
    sites:
      - name: DN.se
        url: https://dn.se
      - name: Feber
        url: https://feber.se
      - name: Hacker news
        url: https://news.ycombinator.com/
      - name: Verge
        url: https://www.theverge.com/
      - name: Engadget
        url: https://www.engadget.com/
  - name: Search
    icon: fa-search
    sites:
      - name: Google
        url: https://google.com
      - name: Duck duck go
        url: https://duckduckgo.com/
      - name: Bing
        url: https://bing.com
  - name: Home automation
    icon: fa-server
    sites:
      - name: Docker
        url: https://www.docker.com/
      - name: Home assistant
        url: https://www.home-assistant.io/

```

## Run

To run this with a persisted config (and favicons) you can run it like this. 

```bash
docker run -v ${pwd}/config.yml:/config.yml -v ${pwd}/favicons:/favicons -p 8080:8080 gosh
```

## 
