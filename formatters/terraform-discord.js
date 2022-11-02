
const data = JSON.parse(input);

const field = {
    name: data.notifications[0].message,
    value: data.notifications[0].run_updated_at,
}

const colors = {
    created: 0xd7aef2,
    applied: 0x66e3a2,
    errored: 0xe34958,
    canceled: 0xf0c18b,
    running: 0x4989e3,
}

const getColor = (status) => {
    const keys = Object.keys(colors)
    for(let i = 0; i < keys.length; i++) {
        const key = keys[i]
        if(status.toLowerCase().includes(key)) {
            return colors[key];
        }
    }
    return colors.running;
}

output = JSON.stringify({
    username: `${data.organization_name} / ${data.workspace_name}`,
    content: ``,
    embeds: [
        {
            author: {
                name: data.run_id,
            },
            title: data.run_message,
            url: data.run_url,
            description: field.name,
            color: getColor(field.name),
            footer: {
                text: `${data.run_created_by} - ${field.value}`,
            }
        }
    ]
})
